/**
 * Возвращает примитивы Web Crypto, которых браузер не даёт вне secure context.
 *
 * `crypto.subtle` и `crypto.randomUUID` помечены в спецификации как
 * [SecureContext]: по https и на localhost они есть, по http на IP или домене —
 * отсутствуют вовсе, даже не как undefined-поле. Приложению это ломает две
 * вещи:
 *
 *   - вход: oidc-client-ts считает PKCE code_challenge через
 *     crypto.subtle.digest("SHA-256") и без него бросает исключение до того,
 *     как дойдёт до редиректа на провайдера;
 *   - события онбординга: они помечаются идентификатором из
 *     crypto.randomUUID().
 *
 * Модуль подставляет обе функции, когда их нет, и ничего не делает, когда они
 * есть, — на https и localhost выполняется штатная реализация браузера.
 *
 * ВАЖНО, чего этот модуль НЕ делает: он не превращает http в безопасный
 * транспорт. PKCE-challenge считается настоящий (тот же SHA-256), но по http
 * токен и код авторизации идут открытым текстом, и любой на пути их прочитает.
 * Это осознанный размен для стенда, а не замена https.
 */

const SHA256_K = new Uint32Array([
  0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5, 0x3956c25b, 0x59f111f1,
  0x923f82a4, 0xab1c5ed5, 0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3,
  0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174, 0xe49b69c1, 0xefbe4786,
  0x0fc19dc6, 0x240ca1cc, 0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
  0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7, 0xc6e00bf3, 0xd5a79147,
  0x06ca6351, 0x14292967, 0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13,
  0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85, 0xa2bfe8a1, 0xa81a664b,
  0xc24b8b70, 0xc76c51a3, 0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
  0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5, 0x391c0cb3, 0x4ed8aa4a,
  0x5b9cca4f, 0x682e6ff3, 0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208,
  0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2,
]);

function rotr(value: number, bits: number): number {
  return ((value >>> bits) | (value << (32 - bits))) >>> 0;
}

/** Чистая реализация SHA-256 (FIPS 180-4) над байтами. */
export function sha256(input: Uint8Array): Uint8Array {
  const state = new Uint32Array([
    0x6a09e667, 0xbb67ae85, 0x3c6ef372, 0xa54ff53a, 0x510e527f, 0x9b05688c,
    0x1f83d9ab, 0x5be0cd19,
  ]);

  // Сообщение дополняется битом 1, нулями и 64-битной длиной в битах так,
  // чтобы итог делился на 64 байта.
  const padded = new Uint8Array(((((input.length + 8) >> 6) + 1) << 6));
  padded.set(input);
  padded[input.length] = 0x80;

  const view = new DataView(padded.buffer);
  const bitLength = input.length * 8;
  view.setUint32(padded.length - 8, Math.floor(bitLength / 0x100000000), false);
  view.setUint32(padded.length - 4, bitLength >>> 0, false);

  const w = new Uint32Array(64);

  for (let offset = 0; offset < padded.length; offset += 64) {
    for (let i = 0; i < 16; i++) {
      w[i] = view.getUint32(offset + i * 4, false);
    }
    for (let i = 16; i < 64; i++) {
      const s0 = rotr(w[i - 15], 7) ^ rotr(w[i - 15], 18) ^ (w[i - 15] >>> 3);
      const s1 = rotr(w[i - 2], 17) ^ rotr(w[i - 2], 19) ^ (w[i - 2] >>> 10);
      w[i] = (w[i - 16] + s0 + w[i - 7] + s1) >>> 0;
    }

    let [a, b, c, d, e, f, g, h] = state;

    for (let i = 0; i < 64; i++) {
      const s1 = rotr(e, 6) ^ rotr(e, 11) ^ rotr(e, 25);
      const ch = (e & f) ^ (~e & g);
      const temp1 = (h + s1 + ch + SHA256_K[i] + w[i]) >>> 0;
      const s0 = rotr(a, 2) ^ rotr(a, 13) ^ rotr(a, 22);
      const maj = (a & b) ^ (a & c) ^ (b & c);
      const temp2 = (s0 + maj) >>> 0;

      h = g;
      g = f;
      f = e;
      e = (d + temp1) >>> 0;
      d = c;
      c = b;
      b = a;
      a = (temp1 + temp2) >>> 0;
    }

    state[0] = (state[0] + a) >>> 0;
    state[1] = (state[1] + b) >>> 0;
    state[2] = (state[2] + c) >>> 0;
    state[3] = (state[3] + d) >>> 0;
    state[4] = (state[4] + e) >>> 0;
    state[5] = (state[5] + f) >>> 0;
    state[6] = (state[6] + g) >>> 0;
    state[7] = (state[7] + h) >>> 0;
  }

  const digest = new Uint8Array(32);
  const digestView = new DataView(digest.buffer);
  for (let i = 0; i < 8; i++) {
    digestView.setUint32(i * 4, state[i], false);
  }
  return digest;
}

/** UUID v4 из crypto.getRandomValues — он доступен и вне secure context. */
export function randomUUIDv4(): string {
  const bytes = new Uint8Array(16);
  crypto.getRandomValues(bytes);

  // Версия и вариант по RFC 4122.
  bytes[6] = (bytes[6] & 0x0f) | 0x40;
  bytes[8] = (bytes[8] & 0x3f) | 0x80;

  const hex: string[] = [];
  for (const byte of bytes) {
    hex.push(byte.toString(16).padStart(2, "0"));
  }
  return [
    hex.slice(0, 4).join(""),
    hex.slice(4, 6).join(""),
    hex.slice(6, 8).join(""),
    hex.slice(8, 10).join(""),
    hex.slice(10, 16).join(""),
  ].join("-");
}

function toBytes(data: BufferSource): Uint8Array {
  if (data instanceof ArrayBuffer) {
    return new Uint8Array(data);
  }
  return new Uint8Array(data.buffer, data.byteOffset, data.byteLength);
}

function normalizeAlgorithm(algorithm: AlgorithmIdentifier): string {
  return typeof algorithm === "string" ? algorithm : algorithm.name;
}

/**
 * Ставит недостающие примитивы. Идемпотентна и безопасна к повторному вызову;
 * возвращает список того, что реально пришлось подставить, — это удобно и для
 * отладки на стенде.
 *
 * target вынесен в параметр ради тестов: в insecure context браузер не
 * объявляет эти поля вообще, а в jsdom они живут на прототипе Crypto, и удалить
 * их у экземпляра нельзя. Подменять прототип в тестах — значит проверять не то,
 * что выполняется в бою.
 */
export function installInsecureContextCrypto(
  target: Crypto = globalThis.crypto,
): string[] {
  const installed: string[] = [];

  if (!target) {
    return installed;
  }

  if (!target.subtle) {
    // Только digest: остальные операции SubtleCrypto приложению не нужны, и
    // молча подсовывать их урезанные версии было бы хуже, чем не иметь вовсе —
    // вызов упадёт с внятным "is not a function", а не с тихо неверным
    // результатом.
    const subtle = {
      digest(algorithm: AlgorithmIdentifier, data: BufferSource): Promise<ArrayBuffer> {
        const name = normalizeAlgorithm(algorithm).toUpperCase();
        if (name !== "SHA-256") {
          return Promise.reject(
            new Error(`insecureContextCrypto: поддерживается только SHA-256, запрошен ${name}`),
          );
        }
        const digest = sha256(toBytes(data));
        // Копия в свежий ArrayBuffer: у Uint8Array поле buffer типизировано как
        // ArrayBufferLike, а SubtleCrypto.digest обязан вернуть именно
        // ArrayBuffer.
        const result = new ArrayBuffer(digest.byteLength);
        new Uint8Array(result).set(digest);
        return Promise.resolve(result);
      },
    };

    Object.defineProperty(target, "subtle", {
      value: subtle,
      configurable: true,
      enumerable: false,
      writable: false,
    });
    installed.push("crypto.subtle.digest");
  }

  if (!target.randomUUID) {
    Object.defineProperty(target, "randomUUID", {
      value: randomUUIDv4,
      configurable: true,
      enumerable: false,
      writable: false,
    });
    installed.push("crypto.randomUUID");
  }

  return installed;
}
