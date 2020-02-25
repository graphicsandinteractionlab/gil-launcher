#!/usr/bin/env python3

from flask import Flask, render_template
import subprocess

app = Flask(__name__)


def launch_browser():
#    res = subprocess.run(['firefox','--kiosk','http://localhost:5000'], capture_output=True)
#   print(res)
    subprocess.Popen(['firefox','--kiosk','http://localhost:5000'])



@app.route('/')
def index():
    return render_template('view.html', my_string="Wheeeee!", my_list=[0,1,2,3,4,5])

if __name__ == '__main__':
    launch_browser()
    app.run(debug=True)
