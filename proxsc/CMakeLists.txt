cmake_minimum_required(VERSION 3.13.2)
project(proxsc)

set(CMAKE_CXX_STANDARD 17)

set(Boost_USE_STATIC_LIBS ON)
set(Boost_USE_MULTITHREAD ON)
find_package(Boost 1.42.0 REQUIRED COMPONENTS system thread regex)
IF(Boost_FOUND)
  include_directories(${Boost_INCLUDE_DIRS})
  link_directories(${Boost_LIBRARY_DIRS})
ENDIF(Boost_FOUND)

set(USED_LIBS ${Boost_SYSTEM_LIBRARY} ${Boost_THREAD_LIBRARY} ${Boost_REGEX_LIBRARY})

add_executable(
    proxsc src/main.cpp
    src/generator.h src/generator.cpp
    src/poller.cpp src/poller.h
    src/session.h src/session.cpp
    src/utils.cpp src/utils.h
)
target_link_libraries(proxsc ${USED_LIBS})
