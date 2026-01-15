@echo off

FOR /f "tokens=*" %%i IN ('minikube docker-env --shell cmd') DO %%i

docker image ls >nul 2>&1
IF %ERRORLEVEL% NEQ 0 (
    echo ? Error: Cannot connect to Docker inside Minikube.
    pause
    exit /b
)

docker build -t game-lobby:0.1 ./game-lobby
docker build -t game-director:0.1 ./game-director
docker build -t game-mmf:0.1 ./game-mmf
docker build -t simple-game-server:0.1 ./simple-game-server

kubectl apply -f simple-game-server/fleet.yaml
kubectl apply -f simple-game-server/fleet-autoscaler.yaml
kubectl apply -f . --recursive

kubectl delete pod -l app=game-lobby
kubectl delete pod -l app=game-director
kubectl delete pod -l app=my-mmf

kubectl get fleet
pause
