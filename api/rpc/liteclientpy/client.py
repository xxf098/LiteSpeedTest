import grpc
import lite_pb2_grpc as pb2_grpc
import lite_pb2 as pb2


class LiteClient(object):
    """
    Client for gRPC functionality
    """

    def __init__(self):
        self.host = '127.0.0.1'
        self.server_port = 10999

        # instantiate a channel
        self.channel = grpc.insecure_channel(
            '{}:{}'.format(self.host, self.server_port))

        # bind the client and the server
        self.stub = pb2_grpc.TestProxyStub(self.channel)

    def start_test(self):
        """
        Client function to call the rpc for StartTest
        """
        message = pb2.TestRequest(
            GroupName="Default",
            SpeedTestMode=pb2.SpeedTestMode.all,
            PingMethod=pb2.PingMethod.googleping,
            SortMethod=pb2.SortMethod.rspeed,
            Concurrency=2,
            TestMode=2,
            Subscription="https://raw.githubusercontent.com/freefq/free/master/v2",
            Language="en",
            FontSize=24,
            Theme="rainbow",
            Timeout=10,
            OutputMode=0
            )
        print(f'{message}')
        for response in self.stub.StartTest(message):
            print(f'{response}')


if __name__ == '__main__':
    client = LiteClient()
    client.start_test()
  