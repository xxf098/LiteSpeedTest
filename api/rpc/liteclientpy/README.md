### setup environment
```bash
# python3 -m pip install virtualenv
virtualenv -p python3 env
source env/bin/activate
pip install grpcio grpcio-tools
cp ../lite/lite.proto ./
# generate lite_pb2.py lite_pb2_grpc.py
python3 -m grpc_tools.protoc --proto_path=. ./lite.proto --python_out=. --grpc_python_out=.
```

### start test
```bash
# open a new terminal to start the grpc server
./lite -grcp -p 10999

# open another terminal to start the python client
python3 client.py
```

