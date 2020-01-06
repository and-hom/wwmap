export function NumberCoordMarshaller(degreeDigits, fractionalPartLength) {
    this.degreeDigits = degreeDigits;
    this.fractionalPartLength = fractionalPartLength;
    this.fractMult = Math.pow(10, fractionalPartLength);
}

NumberCoordMarshaller.prototype.marshal = function (val) {
    let sign = val < 0 ? "-" : "";
    let absVal = Math.abs(val);
    let intVal = Math.floor(absVal);
    let floatPart = absVal - intVal;
    let floatPartInt = withLeadingZero(Math.floor(floatPart * this.fractMult), this.fractionalPartLength);
    return sign + withLeadingZero(intVal, this.degreeDigits) + floatPartInt;
};

NumberCoordMarshaller.prototype.unmarshal = function (txt) {
    if (!txt || txt.trim() == "") {
        return 0.0;
    }
    return parseFloat(txt) / Math.pow(10, txt.length - this.degreeDigits - (txt.startsWith('-') ? 1 : 0));
};



export function DegMinCoordMarshaller(degreeDigits, signPlus, signMinus) {
    this.degreeDigits = degreeDigits;
    this.signPlus = signPlus;
    this.signMinus = signMinus;
}

DegMinCoordMarshaller.prototype.marshal = function (val) {
    let sign = val > 0 ? this.signPlus : this.signMinus;
    let abs = Math.abs(val);

    let degInt = Math.floor(abs);
    let deg = withLeadingZero(degInt, this.degreeDigits);

    let min = Math.floor((abs - degInt) * 6000000) / 100000.0;
    return sign + deg + min;
};

DegMinCoordMarshaller.prototype.unmarshal = function (txt) {
    let k = txt.startsWith(this.signPlus) ? 1 : -1;
    let degStr = txt.substring(1, this.degreeDigits + 1);
    let minStr = txt.substring(this.degreeDigits + 1);

    let deg = degStr == "" ? 0 : parseInt(degStr);
    let min = minStr == "" ? 0 : parseInt(minStr);

    if (minStr.length == 1) {
        min *= 10.0;
    } else if (minStr.length > 2) {
        min /= Math.pow(10, (minStr.length - 2))
    }

    return k * (deg + min / 60);
};



export function DegMinSecCoordMarshaller(degreeDigits, signPlus, signMinus) {
    this.degreeDigits = degreeDigits;
    this.signPlus = signPlus;
    this.signMinus = signMinus;
}

DegMinSecCoordMarshaller.prototype.marshal = function (val) {
    let sign = val > 0 ? this.signPlus : this.signMinus;
    let abs = Math.abs(val);

    let degInt = Math.floor(abs);
    let deg = withLeadingZero(degInt, this.degreeDigits);

    let minFloat = (abs - degInt) * 60;
    let minInt = Math.floor(minFloat);
    let min = withLeadingZero(minInt, 2);

    let sec = Math.floor((minFloat - minInt) * 60000) / 1000.0;
    return sign + deg + min + sec;
};

DegMinSecCoordMarshaller.prototype.unmarshal = function (txt) {
    let k = txt.startsWith(this.signPlus) ? 1 : -1;
    let degStr = txt.substring(1, this.degreeDigits + 1);
    let minStr = txt.substring(this.degreeDigits + 1, this.degreeDigits + 3);
    let secStr = txt.substring(this.degreeDigits + 3);

    let deg = degStr == "" ? 0 : parseInt(degStr);
    let min = minStr == "" ? 0 : parseInt(minStr);
    let sec = secStr == "" ? 0 : parseInt(secStr);

    if (secStr.length == 1) {
        sec *= 10.0;
    } else if (secStr.length > 2) {
        sec /= Math.pow(10, (secStr.length - 2))
    }

    return k * (deg + min / 60 + sec / 3600);
};



function withLeadingZero(val, len) {
    val = "" + val;
    if (val.length < len) {
        return "0".repeat(len - val.length) + val;
    }
    return val
}

export function norm(val, min, max) {
    if (val < min) {
        return min;
    }
    if (val > max) {
        return max;
    }
    return val;
}