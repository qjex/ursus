#include "generator.h"
#include "poller.h"
#include <boost/asio.hpp>
#include <fstream>
#include <iostream>

std::vector<std::string> read_cidrs(std::string filename)
{
    std::vector<std::string> res;
    std::ifstream f(filename.data());
    std::string cidr;
    while (f >> cidr) {
        res.push_back(std::move(cidr));
    }
    return res;
}

int main()
{
    generator g(read_cidrs("cidrs"));
    boost::asio::io_context io_context;
    poller p(io_context, g);
    p.run();
    return 0;
}
