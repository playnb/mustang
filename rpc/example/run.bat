
go build common\mustang\rpc\protoc-gen-gorpc
::--gogofast_out=plugins=grpc
set protocPath=C:\code\server\trunk\branch_0\src\common\mustang\rpc\example\

%protocPath%\protoc.exe  --gorpc_out=plugins=grpc+gorpc,C:\code\server\trunk\branch_0\src\common\mustang\rpc\example\testrpc echo.proto

%protocPath%\protoc.exe  --gogofast_out=plugins=grpc,own_import_prefix=common/protocol/,C:\code\server\trunk\branch_0\src\common\mustang\rpc\example\testrpc2 echo.proto
