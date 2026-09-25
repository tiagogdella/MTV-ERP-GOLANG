# Deploy manual no cluster k3s da empresa

> Sem CD automático ainda — deploy é manual, passo a passo, documentado aqui. Automatizar isso é over-engineering pro estágio atual do projeto (poucos serviços, poucos deploys).

## Acesso ao cluster (configuração única, por máquina de desenvolvimento)

O cluster é um **k3s** rodando no servidor `tiagoserver` (`192.168.1.44`), na rede local.

1. No servidor, copiar o kubeconfig pra um lugar legível pelo seu usuário (ele é root-only por padrão):
   ```bash
   sudo cp /etc/rancher/k3s/k3s.yaml /tmp/k3s-config-tmp.yaml
   sudo chown <seu-usuario>:<seu-usuario> /tmp/k3s-config-tmp.yaml
   ```
2. Na máquina de dev, copiar o arquivo (nunca cola o conteúdo em chat/lugar público — é credencial de acesso total ao cluster):
   ```bash
   mkdir -p ~/.kube
   scp <usuario>@192.168.1.44:/tmp/k3s-config-tmp.yaml ~/.kube/config
   chmod 600 ~/.kube/config
   ```
3. No servidor, apagar a cópia temporária: `rm /tmp/k3s-config-tmp.yaml`
4. No arquivo `~/.kube/config` copiado, trocar `server: https://127.0.0.1:6443` por `server: https://192.168.1.44:6443` (o arquivo original aponta pra localhost, só funciona rodando dentro do próprio servidor).
5. Confirmar: `kubectl get nodes` deve mostrar `tiagoserver Ready control-plane`.

⚠️ **Pegadinha conhecida**: se `kubectl` reclamar de `current-context must exist` ou tentar conectar em `localhost:8080`, confere se não sobrou espaço em branco no início do arquivo (`cat -A ~/.kube/config | head -1` — não pode ter espaço antes de `apiVersion`). Isso quebra o parser de YAML silenciosamente.

## Registro de imagens (ghcr.io)

*(desde 2026-09-14 — antes as imagens eram importadas direto no containerd do node, sem registro. Trocado pelo GitHub Container Registry pra parar de repetir `docker save`/`scp`/`import` a cada deploy de cada serviço.)*

Cada imagem é **privada**, em `ghcr.io/tiagogdella/<nome-do-servico>`. Configuração única, feita uma vez:

1. **Gerar um Personal Access Token (classic)** em `github.com/settings/tokens`, com escopo `write:packages` (já inclui leitura). O GitHub só mostra o token uma vez — guardar num cofre de senhas, nunca em arquivo do repo.
2. **Logar o Docker da máquina de dev**, pra poder dar `push`:
   ```bash
   docker login ghcr.io -u tiagogdella
   ```
   (senha = o token, colado só no prompt interativo — nunca no comando)
3. **Dar credencial pro cluster puxar** — o k3s usa containerd, não o Docker daemon, então o `docker login` acima não vale pra ele. Precisa de um `Secret` do tipo `dockerconfigjson`, criado uma vez **por namespace**:
   ```bash
   read -s GHCR_PAT
   # cola o token, Enter
   kubectl create secret docker-registry ghcr-secret -n mtv-erp \
     --docker-server=ghcr.io \
     --docker-username=tiagogdella \
     --docker-password="$GHCR_PAT" \
     --docker-email=<seu-email>
   unset GHCR_PAT
   ```
   Usar a variável em vez do token direto no comando evita que ele fique gravado no histórico do bash (que guarda o texto digitado, não o valor substituído).

## Deploy de um serviço (repete a cada mudança de código)

Usando `service-template` como exemplo — o mesmo fluxo vale pra qualquer serviço.

1. **Build e push da imagem**, dentro da pasta do serviço:
   ```bash
   docker build --provenance=false --sbom=false --platform=linux/amd64 -t ghcr.io/tiagogdella/<nome-do-servico>:latest .
   docker push ghcr.io/tiagogdella/<nome-do-servico>:latest
   ```
   ⚠️ `--provenance=false --sbom=false --platform=linux/amd64`: o Docker moderno (buildx) gera por padrão uma *manifest list* multi-plataforma com metadados de proveniência. Isso funciona bem no `push`, mas dava problema no fluxo antigo de `import` direto no containerd — mantido aqui por segurança e consistência entre serviços.

2. **Aplicar os manifests** (primeira vez, ou quando o `deployment.yaml`/`service.yaml`/etc. mudou) ou **forçar novo rollout** (quando só a imagem mudou, mesma tag `:latest`) — na máquina de dev:
   ```bash
   kubectl apply -f deploy/<nome-do-servico>/configmap.yaml -f deploy/<nome-do-servico>/deployment.yaml -f deploy/<nome-do-servico>/service.yaml -f deploy/<nome-do-servico>/service-metrics.yaml
   # ou, se os manifests não mudaram e só a imagem foi atualizada:
   kubectl rollout restart deployment/<nome-do-servico> -n mtv-erp
   kubectl rollout status deployment/<nome-do-servico> -n mtv-erp
   ```
   ⚠️ Com tag `:latest`, o Kubernetes **não percebe sozinho** que a imagem mudou — por isso o `rollout restart` é necessário mesmo com `imagePullPolicy: Always`. Isso some quando o CI/CD passar a taggear imagens com hash de commit.

   ⚠️ **Nunca dar `kubectl apply -f deploy/<nome-do-servico>/` na pasta inteira** — cada serviço tem um `secret.yaml` que é só um **molde de exemplo** (senha literal `SENHA`, nunca preenchida de verdade), documentado dentro do próprio arquivo. Aplicar a pasta toda aplica esse molde também e **sobrescreve o Secret real** criado na mão (`kubectl create secret ...`), quebrando a conexão com o banco (`password authentication failed`). Foi o que aconteceu no deploy do inventory-service em 2026-09-23. Lista os arquivos explícitos (como no comando acima) ou aplica um por um, sempre pulando o `secret.yaml`.

3. **Testar** (o `Service` é `ClusterIP`, só acessível de dentro do cluster):
   ```bash
   kubectl port-forward -n mtv-erp svc/<nome-do-servico> 8080:8080
   curl http://localhost:8080/healthz
   ```

### No `deployment.yaml` de um serviço novo

```yaml
    spec:
      containers:
        - name: <nome-do-servico>
          image: ghcr.io/tiagogdella/<nome-do-servico>:latest
          imagePullPolicy: Always
          # ... resto do container (ports, envFrom, probes)
      imagePullSecrets:
        - name: ghcr-secret
```

- `imagePullPolicy: Always` — obrigatório. Com um registro de verdade, o k3s deve **checar** a cada pull em vez de confiar numa cópia local desatualizada.
- `imagePullSecrets` — vai no nível do **Pod** (irmão de `containers:`, não dentro de um item da lista) — erro comum de indentação.

## Coisas que precisa saber antes de deployar em cluster compartilhado

- O cluster já roda um sistema em produção de verdade, no namespace `comprassularroz` — **nunca mexer nesse namespace** (isso inclui o `Secret` `ghcr-secret` de lá, que é de outro projeto — o nosso é o do namespace `mtv-erp`, criado à parte). Nosso projeto vive isolado no namespace `mtv-erp`.
- Segredos (senha de banco, token do ghcr.io, etc.) nunca vão em arquivo `.yaml` commitado — sempre `kubectl create secret ... --from-literal=...` ou `kubectl create secret docker-registry ...`, direto no cluster.
