helm repo add argo https://argoproj.github.io/argo-helm
helm repo update

# 2. 설치 (argocd 네임스페이스에)
helm upgrade --install argocd argo/argo-cd ^
  --namespace argocd ^
  --create-namespace ^
  --set server.service.type=LoadBalancer ^  # UI 접속용 (IP 할당)
  --wait