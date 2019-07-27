#ifndef POLLER_H
#define POLLER_H

#include "generator.h"
#include "session.h"
#include <boost/asio.hpp>
#include <iostream>
#include <unordered_map>

class poller {
public:
    poller(boost::asio::io_context& io_context, const generator& gen);
    void run(size_t);
    void add_nxt_connection();
    void check_deadline(boost::asio::ip::tcp::socket* socket_ptr, const boost::system::error_code& error);
    void end_session(boost::asio::ip::tcp::socket* socket_ptr);

private:
    boost::asio::io_context& io_context;
    generator gen;
    generator::iterator gen_it;
    std::unordered_map<boost::asio::ip::tcp::socket*, std::unique_ptr<session>> sessions;
    std::unordered_map<boost::asio::ip::tcp::socket*, std::unique_ptr<boost::asio::ip::tcp::socket>> sockets;
    std::unordered_map<boost::asio::ip::tcp::socket*, std::unique_ptr<boost::asio::steady_timer>> timers;
};

#endif // POLLER_H
