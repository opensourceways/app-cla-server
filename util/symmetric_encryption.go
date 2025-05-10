/*
 * Copyright (C) 2021. Huawei Technologies Co., Ltd. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

type SymmetricEncryption interface {
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(ciphertext []byte) ([]byte, error)
}

func NewSymmetricEncryption(key, nonce string) (SymmetricEncryption, error) {
	c, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	se := symmetricEncryption{aead: gcm}

	if nonce != "" {
		nonce1, err := hex.DecodeString(nonce)
		if err != nil {
			return nil, err
		}
		if len(nonce1) != gcm.NonceSize() {
			return nil, fmt.Errorf("the length of nonce for symmetric encryption is unmatched")
		}
		se.nonce = nonce1
	}

	return se, nil
}

type symmetricEncryption struct {
	aead  cipher.AEAD
	nonce []byte
}

func (se symmetricEncryption) Encrypt(plaintext []byte) ([]byte, error) {
	nonce := se.nonce
	if nonce == nil {
		nonce = make([]byte, se.aead.NonceSize())
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			return nil, err
		}
	}

	return se.aead.Seal(nonce, nonce, plaintext, nil), nil
}

func (se symmetricEncryption) Decrypt(ciphertext []byte) ([]byte, error) {
	nonceSize := se.aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return se.aead.Open(nil, nonce, ciphertext, nil)
}
