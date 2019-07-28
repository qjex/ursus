#ifndef UTILS_H
#define UTILS_H
#include <boost/asio.hpp>
#include <cstring>
#include <iostream>

namespace utils {
std::string to_string(const boost::asio::ip::tcp::socket& socket);
}

#endif // UTILS_H
