#include "poller.h"
#include "session.h"
#include "utils.h"

poller::poller(boost::asio::io_context& io_context, const generator& gen)
    : io_context(io_context)
    , gen(gen)
    , gen_it(this->gen.begin())

{
}

void poller::run(size_t cap)
{
    for (size_t i = 0; i < cap && gen_it != gen.end(); ++i) {
        add_nxt_connection();
    }
    io_context.run();
}

void poller::add_nxt_connection()
{
    if (gen_it == gen.end()) {
        return;
    }
    using namespace boost::asio;
    ip::tcp::endpoint endpoint(ip::address::from_string(*gen_it), 1080);
    std::unique_ptr<ip::tcp::socket> socket_ptr = std::make_unique<ip::tcp::socket>(io_context);
    std::unique_ptr<steady_timer> timer_ptr = std::make_unique<steady_timer>(io_context);
    auto& t = *timer_ptr;
    auto* s = socket_ptr.get();
    sockets[s] = std::move(socket_ptr);
    timers[s] = std::move(timer_ptr);

    t.expires_after(std::chrono::seconds(2));
    t.async_wait(std::bind(&poller::check_deadline, this, s, *gen_it, std::placeholders::_1));
    std::cerr << "adding: " << (*gen_it) << std::endl;
    s->async_connect(endpoint, [s, this](const boost::system::error_code& error) {
        if (error) {
            if (error.message()[0] == 'O' || error.message()[0] == 'C') {
                return;
            }
            std::cerr << "Error: " << error.message() << std::endl;
            return;
        }
        if (!s->is_open()) {
            end_session(s);
            return;
        }
        std::cerr << "Connected: " << utils::to_string(*s) << std::endl;
        timers[s]->cancel();
        sessions[s] = std::make_unique<session>(io_context, s, *this);
        sessions[s]->start();
    });
    ++gen_it;
}

void poller::check_deadline(boost::asio::ip::tcp::socket* socket_ptr, const std::string& host, const boost::system::error_code& error)
{
    timers.erase(socket_ptr);
    if (error) {
        return;
    }
    if (!sessions.count(socket_ptr)) {
        sockets.erase(socket_ptr);
        add_nxt_connection();
    }
}

void poller::end_session(boost::asio::ip::tcp::socket* socket_ptr)
{
    sessions.erase(socket_ptr);
    sockets.erase(socket_ptr);
}
