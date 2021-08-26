#!/usr/bin/env python3
from flask import Flask, request, jsonify, render_template, redirect, session, url_for
import base64
import urllib.request
import urllib.parse

app = Flask(__name__)
app.secret_key = b'_5#y2L"F4Q8z\n\xec]/'

@app.route('/collect', methods=['GET'])
def collect():
    session.pop('data', None)
    try:
        count = int(request.args.get('count'))
    except:
        return render_template('index.html', error="Invalid count")
    if count < 0 or count > 500:
        return render_template('index.html', error="Invalid count")
    try:
        params = urllib.parse.urlencode({'count': count})
        collectUrl = session.get('collectUrl')
        url = "%s/collect?%s" % (collectUrl, params)
        req = urllib.request.Request(url=url)
        with urllib.request.urlopen(req) as f:
            data = f.read()
    except:
        return render_template('index.html', error="Collect service is unavailable")
    session['data'] = data.decode('ascii')
    return render_template('index.html', data=session['data'])

@app.route('/plot', methods=['POST'])
def plot():
    data = session.get('data')
    if not data:
        return render_template('index.html', error="No data collected")
    bindata = data.encode('ascii')
    try:
        req = urllib.request.Request(url="%s/plot" % session.get('plotUrl'),
                headers={'Content-Type': 'application/json'}, data=bindata)
        with urllib.request.urlopen(req) as f:
            img = f.read()
    except:
        return render_template('index.html', data=data, error="Plot service is unavailable")
    imgplot = base64.b64encode(img).decode('ascii')
    return render_template('index.html', data=data, imgplot=imgplot)

@app.route('/', methods=['GET', 'POST'])
def braine():
    data = session.get('data')
    session['collectUrl'] = request.form.get('collectUrl', '')
    session['plotUrl'] = request.form.get('plotUrl', '')
    return render_template('index.html', data=data)

@app.route('/logout')
def logout():
    session.pop('data', None)
    session.pop('collectUrl', None)
    session.pop('plotUrl', None)
    return redirect(url_for('braine'))

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000, debug=False)
