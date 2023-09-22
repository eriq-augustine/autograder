package util

import (
    "crypto/sha256"
    "encoding/hex"
)

func Sha256Hex(data []byte) string {
    sha := sha256.Sum256(data);
    return hex.EncodeToString(sha[:]);
}