[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](LICENSE.md)
[![Tests](https://github.com/ozneroL541/corner-cases/actions/workflows/test.yml/badge.svg)](https://github.com/ozneroL541/corner-cases/actions/workflows/test.yml)

# Corner Cases

> [!Warning]
> Do not use this code in real-world implementation.

## Library choice for Ed5519
We used [filippo.io/edwards25519](https://github.com/FiloSottile/edwards25519) for computing scalar multyplication and sum of points over the elliptic curve. The library contains the requested operations and is well mantained.  
Moreover, the mantainer seems to be competent and very expirienced in the topics.

<<<<<<< HEAD
## Choice of the inputs for the Verification
This document details the theoretical justification, computation method, and source citations for each of the six test inputs (INP1 - INP6) used to replicate the Ed25519 verification compatibility matrix. 

**Note on Message Selection:** 
For all test cases, the messages (e.g., `"Message1"`, `"Message2"`) were chosen simply as distinct, arbitrary byte strings. Because the message $M$ is hashed along with $R$ and $A$ to compute the challenge scalar $k = H(R \parallel A \parallel M)$, changing the message guarantees a unique hash and prevents any unintended collisions or caching behaviors between test vectors.

---

## INP1: Standard Valid Signature

*   **Why it was chosen:** To establish a baseline. We need an input that is mathematically perfect and strictly conforms to RFC 8032 to ensure that all six verifiers (VER1 - VER6) accept it (yielding `[v, v, v, v, v, v]`).
*   **How it was computed:** We generated a standard keypair using the provided `KeyGen()` function and signed `"Message1"` using the standard `SigGen()` function. The resulting scalar $S$ is strictly less than the curve order $L$, and the point $R$ is canonically encoded.

---

## INP2: Non-reduced Scalar (S + L)

*   **Why it was chosen:** To test the $C_S$ (Reduced Scalar) check. Permissive libraries implicitly reduce $S \pmod L$ during scalar multiplication, while strict libraries explicitly check if $S > L$ and reject it. Such an $S$ should be rejected, otherwise we allow an attacker to perform (multiple) signature malleability.
*   **How it was computed:** We started by generating a standard, valid signature $(R, S)$ for `"Message2"` using `SigGen()`. We then took the scalar $S$, converted it to a big integer, and mathematically added the main subgroup order $L$ ($2^{252} + 27742317777372353535851937790883648493$). We serialized this new value $S' = S + L$ back into 32 bytes. Because $S' \equiv S \pmod L$, the math still holds, but $S'$ is geometrically out of bounds.
*   **Citation:** 
    *   IACR ePrint 2020/1244, Section 5.1 ("Non-canonical scalars").
    *   Henry de Valence, "It's 25519am", Section *Validation checks* (specifically the $S < L$ requirement).

---

## INP3: Small-Order Public Key ($A = \mathcal{O}$)

*   **Why it was chosen:** To test the $C_{small}$ (Small-Order) check. If an attacker provides a public key $A$ that resides in the small-order torsion subgroup (such as the identity point), they can forge signatures without knowing the private key, unless the verifier explicitly checks that $[8]A \neq \mathcal{O}$. 
Indeed, the sign is bounded to the message iff $[k]A\neq 0$ and the order of a point in the elliptic curve is a divisor of $8L$, so requiring  $[8]A \neq \mathcal{O}$ implies that $[k]A$ won't be 0 (unless $k = 0$).
*   **How it was computed:** We maliciously set the public key $A$ to the curve's identity point $\mathcal{O}$ (encoded as $y=1, x=0$, which is `0100...00` in hex). We then chose a random scalar for $S$. Because $A = \mathcal{O}$, the multiplication $[k]A$ always equals $\mathcal{O}$ regardless of the hash $k$. Therefore, to satisfy the uncofactored equation $[S]B = R + [k]A$, we simply computed $R = [S]B$. 
*   **Citation:** 
    *   IACR ePrint 2020/1244, Section 5.3 ("Small order public keys").
    *   Henry de Valence, "It's 25519am", Section *Validation checks* (Small-order check).

---
=======
## Requirements
Either:
- Go 1.27.3
- Make
or:
- Docker

## Compilation
### Natively
To natively compile this project run:
```sh
make build
```
The executable will be in `bin/app`.
### Docker
Otherwise you can build a docker image with:
```sh
docker build -t ex0601 .
```

## Execution
### Natively
```sh
make run
```
### Docker
After compiling the Docker image:
```sh
docker run ex0601:latest
```
>>>>>>> fb24ae86d42147461d2b61c2f35453c5c270b686

## Author
- 267234 Marco Egidio Pusceddu @MARCO5P
- 267646 Radice Lorenzo @ozneroL541

## License
Copyright (C) 2026 Marco Egidio Pusceddu, Lorenzo Radice

This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program. If not, see http://www.gnu.org/licenses/.
