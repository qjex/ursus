#ifndef SESSION_H
#define SESSION_H
#include <boost/asio.hpp>

class poller;

class session {
public:
    session(boost::asio::io_context& io_context, boost::asio::ip::tcp::socket&& s, poller& p);

    void start();

private:
    boost::asio::io_context& io_context;
    boost::asio::ip::tcp::socket s;
    boost::asio::steady_timer timer;
    poller& p;
};

#endif // SESSION_H
