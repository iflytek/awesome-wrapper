#!/bin/bash

g++ -shared -fPIC -g -Wno-attributes -fvisibility=hidden -std=c++11 -I ../../include/ -o ./libwrapper.so ./wrapper.cpp