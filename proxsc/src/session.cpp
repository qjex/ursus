#include "session.h"
#include "poller.h"
#include "utils.h"

session::session(boost::asio::io_context& io_context, boost::asio::ip::tcp::socket* s, poller& p)
    : io_context(io_context)
    , s(s)
    , timer(std::make_shared<boost::asio::steady_timer>(io_context))
    , p(p)
    , host(utils::to_string(*s))
    , stoped(false)
{
}

void session::start()
{
    state = session_state::SEND_GREETING;
    handle();
}

void session::handle()
{
    std::cerr << utils::to_string(*s) << " state=" << static_cast<int>(state) << std::endl;
    timer->expires_at(boost::asio::steady_timer::time_point::max());
    switch (state) {
    case session_state::SEND_GREETING:
        send_greetings();
        break;
    case session_state::READ_METHOD:
        read_method();
        break;
    case session_state::HANDLE_METHOD:
        handle_method();
        break;
    case session_state::SEND_REQUEST:
        send_request();
        break;
    case session_state::READ_RESPONSE:
        read_response();
        break;
    }
}

void session::send_greetings()
{
    buf[0] = 0x05;
    buf[1] = 0x01;
    buf[2] = 0x00; // NO AUTHENTICATION REQUIRED
    set_deadline(2);
    boost::asio::async_write(*s, boost::asio::buffer(buf, 3), [this](const boost::system::error_code& error, std::size_t size) {
        if (stoped || error || size != 3) {
            close();
            return;
        }
        state = session_state::READ_METHOD;
        handle();
    });
}

void session::read_method()
{
    set_deadline(2);
    boost::asio::async_read(*s, boost::asio::buffer(buf, 2), [this](const boost::system::error_code& error, std::size_t size) {
        if (stoped || error || size != 2) {
            close();
            return;
        }
        state = session_state::HANDLE_METHOD;
        handle();
    });
}

void session::handle_method()
{
    if (buf[0] != 0x05) {
        std::cerr << "Unsupported socks version: " << static_cast<int>(buf[0]) << std::endl;
        close();
    } else if (buf[1] != 0x00) {
        std::cerr << "Unsupported auth method: " << static_cast<int>(buf[1]) << std::endl;
        close();
    } else {
        state = session_state::SEND_REQUEST;
        handle();
    }
}

void session::send_request()
{
    buf[0] = 0x05;
    buf[1] = 0x01;
    buf[2] = 0;
    buf[3] = 0x03;
    buf[4] = 11;
    char addr[11] = "google.com";
    std::copy(std::begin(addr), std::end(addr), std::begin(buf) + 5);
    buf[16] = 0;
    buf[17] = 80;

    set_deadline(2);
    boost::asio::async_write(*s, boost::asio::buffer(buf, 18), [this](const boost::system::error_code& error, std::size_t size) {
        if (stoped || error || size != 18) {
            close();
            return;
        }
        state = session_state::READ_RESPONSE;
        handle();
    });
}

void session::read_response()
{
    set_deadline(4);
    boost::asio::async_read(*s, boost::asio::buffer(buf, 2), [this](const boost::system::error_code& error, std::size_t size) {
        if (stoped || error || size != 2) {
            std::cerr << "error reading response: " << error.message() << std::endl;
            close();
            return;
        }
        if (buf[1] != 0) {
            std::cerr << "NOT OK! " << static_cast<int>(buf[1]) << ' ' << host << std::endl;

        } else {
            std::cerr << "OK! " << host << std::endl;
        }
        close();
    });
}

void session::set_deadline(int secs)
{
    timer->expires_after(std::chrono::seconds(secs));
    timer->async_wait([this, t = timer](const boost::system::error_code& error) {
        if (error) {
            return;
        }
        if (t->expiry() == boost::asio::steady_timer::time_point::max()) {
            std::cerr << "in deadline: " << utils::to_string(*s) << std::endl;
            return;
        }
        //        s->close();
        stoped = true;
    });
}

void session::close()
{
    timer->expires_at(boost::asio::steady_timer::time_point::max());
    p.add_nxt_connection();
    p.end_session(s);
}
