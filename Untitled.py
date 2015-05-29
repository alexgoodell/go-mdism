# -*- coding: utf-8 -*-
"""
Spyder Editor

This is a temporary script file.
"""

text = "hello world"

import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
import csv
with open('master.csv', 'rb') as f:
    reader = csv.reader(f)

df = pd.DataFrame(mastercsv[1:], columns=mastercsv[0])
df = df[df.columns].astype(int)
hivdf = df[df.Model_id == 1]
newhivdf = hivdf.drop('Model_id', axis=1, level=None, inplace=False)
final = newhivdf.pivot(index='Person_id',columns='Cycle_id')

for i in range(1, 6): 
    plt.plot(final.iloc[i])