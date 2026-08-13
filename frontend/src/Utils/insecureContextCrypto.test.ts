import { describe, it, expect } from "vitest";
import {
  sha256,
  randomUUIDv4,
  installInsecureContextCrypto,
} from "./insecureContextCrypto";

function hex(bytes: Uint8Array): string {
  return Array.from(bytes)
    .map((b) => b.toString(16).padStart(2, "0"))
    .join("");
}

describe("sha256", () => {
  // Контрольные векторы FIPS 180-4. Реализация своя, поэтому проверяется она
  // не на «что-то посчиталось», а на точное совпадение с эталоном.
  it("считает хэш пустой строки", () => {
    expect(hex(sha256(new Uint8Array(0)))).toBe(
      "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
    );
  });

  it("считает хэш 'abc'", () => {
    expect(hex(sha256(new TextEncoder().encode("abc")))).toBe(
      "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad",
    );
  });

  it("считает хэш сообщения в два блока", () => {
    const input = "abcdbcdecdefdefgefghfghighijhijkijkljklmklmnlmnomnopnopq";
    expect(hex(sha256(new TextEncoder().encode(input)))).toBe(
      "248d6a61d20638b8e5c026930c3e6039a33ce45964ff2167f6ecedd419db06c1",
    );
  });

  it("считает хэш на границе выравнивания блока", () => {
    // 55 байт помещаются в один блок вместе с padding, 56 — уже нет: это тот
    // самый случай, на котором ошибаются самописные реализации.
    expect(hex(sha256(new TextEncoder().encode("a".repeat(55))))).toBe(
      "9f4390f8d30c2dd92ec9f095b65e2b9ae9b0a925a5258e241c9f1e910f734318",
    );
    expect(hex(sha256(new TextEncoder().encode("a".repeat(56))))).toBe(
      "b35439a4ac6f0948b6d6f9e3c6af0f5f590ce20f1bde7090ef7970686ec6738a",
    );
  });

  it("считает хэш длинного сообщения", () => {
    expect(hex(sha256(new TextEncoder().encode("a".repeat(1000))))).toBe(
      "41edece42d63e8d9bf515a9ba6932e1c20cbc9f5a5d134645adb5db1b9737ea3",
    );
  });
});

describe("randomUUIDv4", () => {
  it("возвращает корректный UUID четвёртой версии", () => {
    const uuid = randomUUIDv4();
    expect(uuid).toMatch(
      /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/,
    );
  });

  it("не повторяется", () => {
    const generated = new Set(Array.from({ length: 500 }, randomUUIDv4));
    expect(generated.size).toBe(500);
  });
});

describe("installInsecureContextCrypto", () => {
  /**
   * Crypto, каким его видит страница по http: getRandomValues есть, а помеченных
   * [SecureContext] subtle и randomUUID нет вовсе. Именно объект, а не
   * подмена глобального crypto: в jsdom эти поля объявлены на прототипе, и
   * удалить их у экземпляра невозможно.
   */
  function insecureCrypto(): Crypto {
    return {
      getRandomValues: <T extends ArrayBufferView | null>(array: T): T =>
        globalThis.crypto.getRandomValues(array as never) as T,
    } as Crypto;
  }

  it("подставляет отсутствующие примитивы", async () => {
    const target = insecureCrypto();

    const installed = installInsecureContextCrypto(target);

    expect(installed).toEqual(["crypto.subtle.digest", "crypto.randomUUID"]);
    expect(typeof target.randomUUID()).toBe("string");

    const digest = await target.subtle.digest(
      "SHA-256",
      new TextEncoder().encode("abc"),
    );
    expect(hex(new Uint8Array(digest))).toBe(
      "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad",
    );
  });

  it("не трогает то, что браузер уже предоставил", () => {
    const marker = { digest: () => Promise.resolve(new ArrayBuffer(0)) };
    const target = {
      subtle: marker,
      randomUUID: () => "fixed",
    } as unknown as Crypto;

    expect(installInsecureContextCrypto(target)).toEqual([]);
    expect(target.subtle).toBe(marker);
    expect(target.randomUUID()).toBe("fixed");
  });

  it("повторный вызов ничего не меняет", () => {
    const target = insecureCrypto();

    expect(installInsecureContextCrypto(target)).toHaveLength(2);
    expect(installInsecureContextCrypto(target)).toEqual([]);
  });

  it("отказывается считать алгоритм, который не поддерживает", async () => {
    const target = insecureCrypto();
    installInsecureContextCrypto(target);

    await expect(
      target.subtle.digest("SHA-512", new Uint8Array(1)),
    ).rejects.toThrow(/только SHA-256/);
  });

  it("на настоящем окружении jsdom ничего не подставляет", () => {
    // Косвенно подтверждает, что на https и localhost модуль — no-op.
    expect(installInsecureContextCrypto()).toEqual([]);
  });
});
