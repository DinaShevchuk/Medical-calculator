package repository

import (
    "dosage_drugs/internal/app/ds"
)

func (r *Repository) RegisterUser(login, password string) (*ds.User, error) {
    u := &ds.User{Login: login, Password: password}
    if err := r.db.Create(u).Error; err != nil {
        return nil, err
    }
    return u, nil
}
