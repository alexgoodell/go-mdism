# -*- coding: utf-8 -*-
"""
Spyder Editor

This is a temporary script file.
"""

text = "hello world"

import pandas as pd
import numpy as np
import matplotlib.pyplot as plt


mastercsv = pd.read_csv('tmp/master.csv')

df = mastercsv
df = df[df.columns].astype(int)
hivdf = df[df.Model_id == 0]
newhivdf = hivdf.drop('Model_id', axis=1, level=None, inplace=False)
final = newhivdf.pivot(index='Person_id',columns='Cycle_id')

for i in final.index: 
    plt.plot(final.iloc[i-1])