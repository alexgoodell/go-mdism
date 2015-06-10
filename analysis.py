# -*- coding: utf-8 -*-
"""
Spyder Editor

This is a temporary script file.
"""

text = "hello world"

import pandas as pd
import numpy as np
import matplotlib.pyplot as plt



os.chdir('/Applications/XAMPP/xamppfiles/htdocs/ghd_model_go/gocode/src/github.com/alexgoodell/go-mdism/tmp/')


df = pd.read_csv('master39463.276246739.csv', header=0, skip_blank_lines=1)

# mastercsv = pd.read_csv('tmp/master.csv')

df = df[df.columns].astype(int)
hivdf = df[df.Model_id == 0]
newhivdf = hivdf.drop('Model_id', axis=1, level=None, inplace=False)
final = newhivdf.pivot_table(index='Person_id',columns='Cycle_id')

for i in final.index: 
    plt.plot(final.iloc[i-1])
    
people_with_nafld =   [0] * 10000
for i in range(0, 19):
    newdfl = df[df.State_id == 3]
    people_with_nafld[i] = len(newdfl[newdfl.Cycle_id == i])
    