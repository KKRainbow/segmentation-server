cd frontend
npm run build
cd -
go build main.go
echo "buil ok"
rm dist -rf
mkdir -p dist/frontend
mv frontend/dist ./dist/frontend/
ErmineLightTrial.x86_64 --output dist/segserver main
scp -r ./dist chenhao@octr:~/SegServer
