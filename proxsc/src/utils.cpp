#include "utils.h"

namespace utils {
std::string to_string(const boost::asio::ip::tcp::socket& socket)
{
    boost::system::error_code ec;
    std::string res = socket.remote_endpoint(ec).address().to_string();
    //    if (ec) {
    //        std::cerr << "Error in to string: " << ec.message() << std::endl;
    //    }
    return res;
}
}
