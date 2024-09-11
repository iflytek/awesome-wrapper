#ifndef __MANAGER_H__
#define __MANAGER_H__

#include <string>

class manager
{
public:
    manager(int status, std::string data);
    ~manager();
    int get_status();
    void set_status(int status);
    std::string get_data();
    void set_data(std::string data);
private:
    std::string data;
    int status;
};

#endif // __MANAGER_H__