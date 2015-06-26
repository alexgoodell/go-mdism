# -*- coding: utf-8 -*-
"""
Created on Fri Jun 26 11:17:56 2015

@author: alexgoodell
"""


import os
import pandas as pd
import numpy as np
import matplotlib.pyplot as plt

os.chdir('/Applications/XAMPP/xamppfiles/htdocs/ghd_model_go/gocode/src/github.com/alexgoodell/go-mdism/stats-tests/gamma')
df = pd.read_csv('gamma.csv', names=["dist"], skip_blank_lines=1)

plt.figure();

df.plot(kind='hist', alpha=0.5)
