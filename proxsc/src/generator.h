#ifndef GENERATOR_H
#define GENERATOR_H

#include <string>
#include <vector>

class generator {

public:
    generator(const std::vector<std::string>& a);
    class iterator {
        friend class generator;

    public:
        iterator& operator++();
        std::string operator*();
        bool operator==(const iterator& rhs);
        bool operator!=(const iterator& rhs);

    private:
        iterator(generator* gen, size_t pos, unsigned int k, long long mx);

    private:
        size_t pos;
        unsigned int k;
        long long mx;
        generator* ptr;
    };

    std::string get(size_t pos, unsigned int k);

    iterator begin();

    iterator end();

private:
    std::pair<unsigned int, short> parse(const std::string&);
    static std::string to_string(unsigned int);

private:
    std::vector<std::pair<unsigned int, short>> data;
};

#endif // GENERATOR_H
