
go build common\mustang\rpc\protoc-gen-gorpc
::--gogofast_out=plugins=grpc
set protocPath=%CD%

%protocPath%\..\example\protoc.exe  --gorpc_out=plugins=grpc+gorpc,%protocPath%\ wire.proto

del protoc-gen-gorpc.exe