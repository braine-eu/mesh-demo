#!/usr/bin/env python3
from flask import Flask, request, send_file
import matplotlib
import matplotlib.pyplot as plt
from io import BytesIO

app = Flask(__name__)

#plt.style.use('seaborn')

@app.route('/plot', methods=['POST'])
def plot():
    fig, ax = plt.subplots()
    x = request.json[0]
    y = request.json[1]
    ax.scatter(x, y)
    #ax.set_title("Random scatterplot")
    img_bytes = BytesIO()
    fig.savefig(img_bytes)
    img_bytes.seek(0)
    response = send_file(img_bytes, mimetype='image/png')
    response.headers['Cache-Control'] = 'no-store, no-cache, must-revalidate, max-age=0'
    return response

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5001, debug=False)
