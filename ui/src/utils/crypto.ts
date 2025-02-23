// @ts-ignore
import CryptoJS from "crypto-js";

const secretKey = 'secret-key-16-ok';

// 加密函数
export function encrypt(message: string) {
    // key 和 iv 使用同一个值
    const sKey = CryptoJS.enc.Utf8.parse(secretKey);
    const encrypted = CryptoJS.AES.encrypt(message, sKey, {
        iv: sKey,
        mode: CryptoJS.mode.CBC, // CBC算法
        padding: CryptoJS.pad.Pkcs7, //使用pkcs7 进行padding 后端需要注意
    });

    return encrypted.toString();
}

// 解密函数
export function decrypt(base64CipherText: string) {
    // key 和 iv 使用同一个值
    const sKey = CryptoJS.enc.Utf8.parse(secretKey);
    const decrypted = CryptoJS.AES.decrypt(base64CipherText, sKey, {
        iv: sKey,
        mode: CryptoJS.mode.CBC, // CBC算法
        padding: CryptoJS.pad.Pkcs7, //使用pkcs7 进行padding 后端需要注意
    });

    return decrypted.toString(CryptoJS.enc.Utf8);
}