# Deploy manual no cluster k3s da empresa

> Sem CD automático ainda — deploy é manual, passo a passo, documentado aqui. Automatizar isso é over-engineering pro estágio atual do projeto (poucos serviços, poucos deploys).

## Acesso ao cluster (configuração única, por máquina de desenvolvimento)

O cluster é um **k3s** rodando no servidor `tiagoserver` (`192.168.1.44`), na rede local. Não existe registro de imagens configurado — como o cluster tem só **1 node**, as imagens são importadas direto nele, sem precisar de registro.

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

## Deploy de um serviço (repete a cada mudança de código)

Usando `service-template` como exemplo — o mesmo fluxo vale pra qualquer serviço.

1. **Build da imagem**, dentro da pasta do serviço:
   ```bash
   docker build -t <nome-do-servico> .
   ```
2. **Exportar e copiar pro servidor**:
   ```bash
   docker save <nome-do-servico>:latest -o <nome-do-servico>.tar
   scp <nome-do-servico>.tar <usuario>@192.168.1.44:~/
   ```
   ⚠️ Copiar pra `~/` (home do usuário), **não pra `/tmp/`** — o SSH desse servidor isola `/tmp` por sessão (`PrivateTmp` do systemd), então um arquivo copiado por `scp` pode não aparecer numa sessão interativa de SSH separada.
3. **Importar no containerd do k3s** (via SSH no servidor):
   ```bash
   ssh <usuario>@192.168.1.44
   sudo k3s ctr images import ~/<nome-do-servico>.tar
   rm ~/<nome-do-servico>.tar
   exit
   ```
4. **Aplicar os manifests** (primeira vez) ou **forçar novo rollout** (quando só a imagem mudou, mesma tag `:latest`) — de volta na máquina de dev:
   ```bash
   kubectl apply -f deploy/<nome-do-servico>/
   # ou, se os manifests não mudaram e só a imagem foi atualizada:
   kubectl rollout restart deployment/<nome-do-servico> -n mtv-erp
   kubectl rollout status deployment/<nome-do-servico> -n mtv-erp
   ```
   ⚠️ Com tag `:latest`, o Kubernetes **não percebe sozinho** que a imagem mudou (o nome/tag do Deployment continua igual) — por isso o `rollout restart` é necessário. Isso some quando o CI/CD passar a taggear imagens com hash de commit.

5. **Testar** (o `Service` é `ClusterIP`, só acessível de dentro do cluster):
   ```bash
   kubectl port-forward -n mtv-erp svc/<nome-do-servico> 8080:8080
   curl http://localhost:8080/healthz
   ```

## Coisas que precisa saber antes de deployar em cluster compartilhado

- O cluster já roda um sistema em produção de verdade, no namespace `comprassularroz` — **nunca mexer nesse namespace**. Nosso projeto vive isolado no namespace `mtv-erp`.
- `imagePullPolicy: IfNotPresent` é obrigatório nos Deployments — sem isso, o Kubernetes tenta baixar a imagem de um registro externo (que não existe pra gente) mesmo já tendo a cópia local importada, e o pod fica em `ImagePullBackOff`.
- Segredos (senha de banco, etc.) nunca vão em arquivo `.yaml` commitado — sempre `kubectl create secret generic ... --from-literal=...`, direto no cluster.
