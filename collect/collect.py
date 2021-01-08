#!/usr/bin/env python3
from flask import Flask, request, jsonify
import random

app = Flask(__name__)

@app.route('/collect', methods=['GET'])
def collect():
    count = int(request.args.get('count'))
    x = [random.randint(0,100) for i in range(count)]
    y = [random.randint(0,100) for i in range(count)]
    return jsonify([x,y])

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5002, debug=False)
