#ifndef SESSION_H
#define SESSION_H
#include <boost/asio.hpp>

class poller;

enum class session_state {
    SEND_GREETING,
    READ_METHOD,
    HANDLE_METHOD,
    SEND_REQUEST,
    READ_RESPONSE
};

class session {
public:
    session(boost::asio::io_context& io_context, boost::asio::ip::tcp::socket* s, poller& p);

    void start();

private:
    void close();
    void handle();
    void send_greetings();
    void read_method();
    void handle_method();
    void send_request();
    void read_response();
    void handle_request();
    void set_deadline(int);

private:
    boost::asio::io_context& io_context;
    boost::asio::ip::tcp::socket* s;
    boost::asio::steady_timer timer;
    poller& p;
    session_state state;
    std::array<unsigned char, 18> buf;
    std::string host;
    bool stoped;
};

#endif // SESSION_H
