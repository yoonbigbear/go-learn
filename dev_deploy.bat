@echo off
echo ========================================================
echo ?? Windows Environment Deployment Start!
echo ========================================================

REM 1. Minikube ?? ?? ??
echo ?? Connecting to Minikube Docker Environment...
FOR /f "tokens=*" %%i IN ('minikube docker-env --shell cmd') DO %%i

REM ?? ??
docker image ls >nul 2>&1
IF %ERRORLEVEL% NEQ 0 (
    echo ? Error: Cannot connect to Docker inside Minikube.
    pause
    exit /b
)

REM 2. ??? ?? ??
echo ?? Building Docker Images...
echo [1/4] Building Game Lobby...
docker build -t game-lobby:0.1 ./game-lobby
echo [2/4] Building Game Director...
docker build -t game-director:0.1 ./game-director
echo [3/4] Building Game MMF...
docker build -t game-mmf:0.1 ./game-mmf
echo [4/4] Building Simple Game Server (Agones)...
docker build -t simple-game-server:0.1 ./simple-game-server

REM 3. YAML ??
echo ?? Applying Kubernetes Manifests...

REM [??] Fleet ??? ??? ?? ???? ?????.
IF EXIST "simple-game-server\fleet.yaml" (
    echo    ?? Found fleet.yaml! Applying...
    kubectl apply -f simple-game-server/fleet.yaml
) ELSE (
    echo    ? Warning: 'simple-game-server\fleet.yaml' not found!
    echo       Please check the file name and path.
)

REM ??? ???? ??
kubectl apply -f . --recursive

REM 4. ?? ???
echo ?? Restarting Pods...
kubectl delete pod -l app=game-lobby
kubectl delete pod -l app=game-director
kubectl delete pod -l app=my-mmf

echo ========================================================
echo ? Deployment Complete!
echo    Check Fleet Status:
kubectl get fleet
echo ========================================================
pause
