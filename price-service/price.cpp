
// g++ -std=c++11 ../proto/price.pb.cc ../proto/price.grpc.pb.cc price.cpp -L/usr/local/lib `pkg-config --cflags protobuf grpc --libs protobuf grpc++ grpc` -Wl,--as-needed -ldl -o price

#include <string>
#include <unordered_map>
#include <grpcpp/grpcpp.h>

#include "../proto/cpp/price.grpc.pb.h"

using grpc::Server;
using grpc::ServerBuilder;
using grpc::ServerContext;
using grpc::Status;
using price::ItemID;
using price::ItemPrice;
using price::Pricer;

const std::unordered_map<int64_t, float> DB = {
  {1, 200},
  {2, 30240},
  {3, 12940},
};

class PriceService final : public Pricer::Service {
  Status GetPrice(ServerContext* context, const ItemID* request, ItemPrice* reply) override {
    auto itemIt = DB.find(request->id());
    if (itemIt != DB.end()) {
      reply->set_price(itemIt->second);
      return Status::OK;
    } else {
      reply->set_price(0);
      return Status::OK;
    }
  }
};

void RunServer() {
  std::string server_address("0.0.0.0:8787");
  PriceService service;

  ServerBuilder builder;
  builder.AddListeningPort(server_address, grpc::InsecureServerCredentials());
  builder.RegisterService(&service);
  std::unique_ptr<Server> server(builder.BuildAndStart());

  server->Wait();
}

int main(int argc, char const *argv[]) {
    RunServer();
    return 0;
}
