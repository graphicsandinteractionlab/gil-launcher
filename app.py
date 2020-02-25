#!/usr/bin/env python3


from flask import Flask
from jinja2 import Template

app = Flask(__name__)





@app.route('/')
def index():
    return 'Hello, World!'

