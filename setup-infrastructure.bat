@echo off
ECHO ========================================================
ECHO Open Match + Agones - Full Setup
ECHO ========================================================

minikube status >nul 2>&1
IF %ERRORLEVEL% NEQ 0 (
    ECHO Minikube is not running. Starting Minikube...
    minikube start --cpus=4 --memory=8192 --ports 7000-7100:7000-7100/udp
) ELSE (
    ECHO Minikube is already running.
)

helm repo add agones https://agones.dev/chart/stable >nul 2>&1
helm repo add open-match https://open-match.dev/chart/stable >nul 2>&1
helm repo update
helm upgrade --install agones agones/agones ^
    --namespace agones-system ^
    --create-namespace ^
    --set gameservers.minPort=7000,gameservers.maxPort=7100 ^
    --wait


helm upgrade --install open-match open-match/open-match ^
    --namespace open-match ^
    --create-namespace ^
    --set open-match-override.enabled=true ^
    --set evaluator.enabled=true ^
    --wait

ECHO.
ECHO ========================================================
ECHO Setup Complete!
ECHO ========================================================
pause
