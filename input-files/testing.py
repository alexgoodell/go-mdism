
# -*- coding: utf-8 -*-
"""
Spyder Editor

This is a temporary script file.
"""

text = "hello world"
print text

import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
import os
import random


os.chdir('/Applications/XAMPP/xamppfiles/htdocs/ghd_model_go/gocode/src/github.com/alexgoodell/go-mdism/input-files/')




df = pd.read_csv('inputs.csv', index_col=1, header=0, skip_blank_lines=1)
df=df.dropna(axis=1,how='all')
outputdf = pd.DataFrame(columns=["from_id","from_name","to_id","to_name"])

for i in df.mode_id.unique():
    for index, row in df[df.mode_id == i].iterrows():
        from_name = row['name']
        from_id = index
        id_to_make_1 = random.randint(1, len(df[df.mode_id == i]))
        x = 1
        for qindex, qrow in df[df.mode_id == i].iterrows():
            to_name = qrow['name']
            to_id = qindex
            if id_to_make_1 == x:
                tp_base = 1
            else:
                tp_base = 0
            x = x + 1
            new_row = {"from_id" : from_id, "from_name" : from_name, "to_id" : to_id, "to_name": to_name, "tp_base": tp_base}
            outputdf = outputdf.append(new_row, ignore_index=True)
            
        
outputdf.to_csv("output.csv")
    

