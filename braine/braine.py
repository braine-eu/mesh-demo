#!/usr/bin/env python
from flask import Flask, request, jsonify, render_template, redirect, session, url_for
import base64
import urllib.request
import urllib.parse

app = Flask(__name__)
app.secret_key = b'_5#y2L"F4Q8z\n\xec]/'

@app.route('/collect', methods=['GET'])
def collect():
    count = int(request.args.get('count'))
    params = urllib.parse.urlencode({'count': count})
    url = "http://localhost:5002/collect?%s" % params
    req = urllib.request.Request(url=url)
    with urllib.request.urlopen(req) as f:
        data = f.read()
    print(f.status)
    session['data'] = data.decode('ascii')
    return redirect(url_for('braine'))

@app.route('/plot', methods=['POST'])
def plot():
    data = session.get('data')
    bindata = data.encode('ascii')
    req = urllib.request.Request(url='http://localhost:5001/plot',
            headers={'Content-Type': 'application/json'}, data=bindata)
    with urllib.request.urlopen(req) as f:
        img = f.read()
    print(f.status)
    imgplot = base64.b64encode(img).decode('ascii')
    return render_template('index.html', data=data, imgplot=imgplot)

@app.route('/')
def braine():
    data = session.get('data')
    return render_template('index.html', data=data)

@app.route('/logout')
def logout():
    session.pop('data', None)
    return redirect(url_for('braine'))

if __name__ == '__main__':
    app.run(host='localhost', port=5000, debug=False)
