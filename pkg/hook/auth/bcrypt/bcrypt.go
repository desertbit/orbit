/*
 *  Juno
 *  Copyright (C) 2016 DesertBit
 */

package bcrypt

import "golang.org/x/crypto/bcrypt"

const (
	Cost = bcrypt.DefaultCost
)

func Generate(password []byte) ([]byte, error) {
	return bcrypt.GenerateFromPassword(password, Cost)
}

func Compare(hashedPassword, password []byte) error {
	return bcrypt.CompareHashAndPassword(hashedPassword, password)
}
