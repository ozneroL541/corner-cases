[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](LICENSE.md)
[![Tests](https://github.com/ozneroL541/corner-cases/actions/workflows/test.yml/badge.svg)](https://github.com/ozneroL541/corner-cases/actions/workflows/test.yml)

# Corner Cases

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

## Library choice for Ed5519
We used [filippo.io/edwards25519](https://github.com/FiloSottile/edwards25519) for computing scalar multyplication and sum of points over the elliptic curve. The library contains the requested operations and is well mantained.  
Moreover, the mantainer seems to be competent and very expirienced in the topics.

## Choice of the inputs for the Verification
This document details the theoretical justification, computation method, and source citations for each of the six test inputs (INP1 - INP6) used to replicate the Ed25519 verification compatibility matrix. 

**Note on Message Selection:** 
For all test cases, the messages (e.g., `"Message1"`, `"Message2"`) were chosen simply as distinct, arbitrary byte strings. Because the message $M$ is hashed along with $R$ and $A$ to compute the challenge scalar $k = H(R \parallel A \parallel M)$, changing the message guarantees a unique hash and prevents any unintended collisions or caching behaviors between test vectors.

---

## INP1: Invalid Signature (Control Case)

*   **Why it was chosen:** To establish a negative baseline. We need an input that is well-formed (canonical $S$, canonical $R$, public key $A$ of large order) but mathematically wrong, so that all six verifiers (VER1 - VER6) reject it (yielding `[x, x, x, x, x, x]`). 
*   **How it was computed:** We generated a standard keypair using the provided `KeyGen()` function and signed `"Message1"` using the standard `SigGen()` function, obtaining a valid signature $(R, S)$. We then tampered with the scalar, computing $S' = S + 1 \pmod L$ and keeping $R$ unchanged. Since $S' < L$, the structural checks (canonical $S$, canonical $R$, small-order $A$) all pass, and only the verification equation can reject it.
*   **Mathematical Proof:** It is enough to show that the cofactored equation fails. The uncofactored equation implies the cofactored one (multiply both sides by 8), so if the cofactored check rejects the signature, the uncofactored one rejects it as well.
    1. $[8][S']B = [8]([S]B + B) = [8](R + [k]A) + [8]B$
    2. This equals $[8](R + [k]A)$ only if $[8]B = \mathcal{O}$.
    3. $B$ has prime order $L$ and $L \nmid 8$, so $[8]B \neq \mathcal{O}$.
    4. Hence the cofactored equation fails, and so does the uncofactored one.
*   **Citation:**
    * IACR ePrint 2020/1244, Introduction.

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

## INP4: Torsion Component added to R
- **Why it was chosen:** To differentiate between the *Uncofactored* and *Cofactored* verification equations. Adding a low-order torsion point $T$ to $R$ breaks the uncofactored equation, but multiplying the whole equation by the cofactor $h=8$ neutralizes $T$ (since $[8]T = \mathcal{O}$), causing cofactored verifiers to accept it. Because multiple valid signatures can be created for the exact same message by adding different torsion points, this edge case formally permits a **signature malleability attack**.
- **How it was computed:** We generated a normal keypair ($sk, A$). Instead of a standard deterministic $R$, we picked a random scalar $r$ and computed $R_{base} = [r]B$. We then selected a torsion point $T$ of order 8 from the `smallOrderPoints` array and computed our malicious $R_{final} = R_{base} + T$. Finally, we computed the challenge hash $k = H(R_{final} \parallel A \parallel M)$ and manually forged the scalar $S = r + k \cdot sk \pmod L$. 
- **Mathematical Proof:** The cofactor equation is perfectly satisfied. Here is the step-by-step expansion:
  1. **Left Side:** 
     $$[8][S]B = [8]([r + k \cdot sk]B)$$
     $$[8][S]B = [8]([r]B + [k][sk]B)$$
     $$[8][S]B = [8](R_{base} + [k]A)$$
  2. **Right Side:** 
     $$[8]R_{final} + [8][k]A = [8](R_{base} + T) + [8][k]A$$
     $$[8]R_{final} + [8][k]A = [8]R_{base} + [8]T + [8][k]A$$
  3. **Neutralizing the Torsion Point:** Because $T$ has an order of 8, $[8]T = \mathcal{O}$ (the identity point).
     $$[8]R_{base} + \mathcal{O} + [8][k]A = [8](R_{base} + [k]A)$$
  4. Both sides resolve to $[8](R_{base} + [k]A)$, proving the equation holds.
- **Citation:** 
  * IACR ePrint 2020/1244, Section 5.4 ("Adding low-order points to R").
  * Henry de Valence, "It's 25519am", Section *The verification equation* (Cofactored vs Uncofactored equations).

---

## INP5: Non-canonical Encoding ($y \ge p$)

-   **Why it was chosen:** To test the $C_{can}$ (Canonical Condition/Check). The prime field for Curve25519 is $p = 2^{255} - 19$. Strict libraries reject any $y$-coordinate $\ge p$, while permissive libraries mathematically reduce it modulo $p$.
-   **How it was computed:** We set $R = \mathcal{O}$ (where the $y$-coordinate is 1). Instead of encoding it canonically as `0100...00`, we added the prime $p$ to the $y$-coordinate. Since $p + 1 = 2^{255} - 18$, this value still fits within the 255-bit representation. We encoded this non-canonical point, computed the challenge $k = H(R_{encoded} \parallel A \parallel M)$, and calculated $S = k \cdot sk \pmod L$ (since $r=0$).
-   **Attack Classification:** This is a **Malicious Signer Malleability Attack** . Unlike scalar malleability (INP2), a third party cannot forge this because altering $R$ changes the hash $k$. However, a malicious signer can intentionally generate this signature to fracture distributed networks, as strict nodes will reject the encoding while permissive nodes will mathematically accept it.
-   **Citation:** 
    *   IACR ePrint 2020/1244, Section 5.2 ("Non-canonical points").
    *   Henry de Valence, "It's 25519am", Section *Validation checks* (Canonical encoding).

---

---

## INP6: Standard Valid Signature

*   **Why it was chosen:** To establish a positive baseline. We need an input that is mathematically perfect and strictly conforms to RFC 8032, so that all six verifiers (VER1 - VER6) accept it (yielding `[v, v, v, v, v, v]`). Together with INP1 (which every verifier rejects), it guarantees that the verifiers are neither trivially rejecting nor trivially accepting every input, and that any difference observed in the other rows is due to the specific corner case being tested.
*   **How it was computed:** We generated a standard keypair using the provided `KeyGen()` function and signed `"Message6"` using the standard `SigGen()` function. The resulting scalar $S$ is strictly less than the curve order $L$, the point $R$ is canonically encoded, and the public key $A$ is not a small-order point (except with negligible probability).
*   **Mathematical Proof:** The verification equation holds exactly, so the cofactored one holds as well:
    1. $[S]B = [r + k \cdot s]B = [r]B + [k]([s]B) = R + [k]A$
    2. Multiplying both sides by 8 preserves the equality, hence also the cofactored equation holds.
    3. All structural checks pass: $S < L$, $R$ and $A$ are canonical, and $[8]A \neq \mathcal{O}$.
*   **Citation:**
    *   RFC 8032, Section 5.1.7 (Verify).

---

---
  
## Author
- 267234 Marco Egidio Pusceddu @MARCO5P
- 267646 Radice Lorenzo @ozneroL541

## License
Copyright (C) 2026 Marco Egidio Pusceddu, Lorenzo Radice

This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program. If not, see http://www.gnu.org/licenses/.
