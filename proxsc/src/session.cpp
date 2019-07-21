#include "session.h"
#include "poller.h"

session::session(boost::asio::io_context& io_context, boost::asio::ip::tcp::socket&& s, poller& p)
    : io_context(io_context)
    , s(std::move(s))
    , timer(io_context)
    , p(p)
{
}

void session::start()
{
    p.add_nxt_connection();
}
