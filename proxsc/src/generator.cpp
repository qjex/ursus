#include "generator.h"
#include <array>
#include <iostream>
#include <string>

generator::generator(const std::vector<std::string>& a)
{
    data.reserve(a.size());
    for (auto& ip : a) {
        data.push_back(parse(ip));
    }
}

std::string generator::get(size_t pos, unsigned int k)
{
    short s = data[pos].second;
    short rem = 32 - s;
    unsigned int mask = static_cast<unsigned int>((((1L << s) - 1) << rem));
    return to_string((data[pos].first & mask) + k);
}

generator::iterator generator::begin()
{
    return iterator(this, 0, 0, (1ll << (32 - data[0].second)) - 1);
}

generator::iterator generator::end()
{
    return iterator(this, data.size(), 0, 0);
}

std::pair<unsigned int, short> generator::parse(const std::string& ip)
{
    size_t it = 0;
    std::string tmp;
    std::pair<unsigned int, short> res = { 0, 0 };
    for (int i = 0; i < 5; i++) {
        while ((ip[it] != '.' && ip[it] != '/') && it < ip.size()) {
            tmp += ip[it++];
        }
        it++;
        if (i != 4) {
            int part = std::stoi(tmp);
            res.first <<= 8;
            res.first |= (part & 255);
        } else {
            res.second = static_cast<short>(std::stoi(tmp));
        }
        tmp = "";
    }
    return res;
}

std::string generator::to_string(unsigned int mask)
{
    std::array<std::string, 4> res;
    for (size_t i = 0; i < 4; i++) {
        res[3 - i] = std::to_string(mask & 255);
        mask >>= 8;
    }

    return res[0] + "." + res[1] + "." + res[2] + "." + res[3];
}

generator::iterator& generator::iterator::operator++()
{
    if (k < mx) {
        k++;
    } else {
        k = 0;
        pos++;
        std::cerr << "new block: " << pos << std::endl;
        if (pos < ptr->data.size()) {
            mx = (1ll << (32 - ptr->data[pos].second)) - 1;
        }
    }
    return *this;
}

std::string generator::iterator::operator*() { return ptr->get(pos, k); }

bool generator::iterator::operator==(const generator::iterator& rhs) { return pos == rhs.pos && k == rhs.k && ptr == rhs.ptr; }

bool generator::iterator::operator!=(const generator::iterator& rhs) { return pos != rhs.pos || k != rhs.k || ptr != rhs.ptr; }

generator::iterator::iterator(generator* gen, size_t pos, unsigned int k, long long mx)
    : pos(pos)
    , k(k)
    , mx(mx)
    , ptr(gen)
{
}
