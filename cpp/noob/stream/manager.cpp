#include "string"
#include "manager.h"

manager::manager(int status, std::string data) {
    this->status = status;
    this->data = data;
}

manager::~manager() {
    // Deallocate memory
}

void manager::set_status(int status) {
    this->status = status;
}

int manager::get_status() {
    return this->status;
}

void manager::set_data(const std::string data) {
    this->data = data;
}