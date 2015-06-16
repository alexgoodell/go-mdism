# -*- coding: utf-8 -*-
"""
Created on Mon Jun 15 22:59:38 2015

@author: alexgoodell
"""

import os
import pandas as pd
import numpy as np
import matplotlib.pyplot as plt


os.chdir('/Applications/XAMPP/xamppfiles/htdocs/ghd_model_go/gocode/src/github.com/alexgoodell/go-mdism/input-files/')
df = pd.read_csv('ras_inputs.csv', header=0, skip_blank_lines=1)

outputdf = pd.DataFrame(columns=["Model_id", "Model_name", "To_state_id",
                                 "To_state_name", "Sex_state_id",
                                 "Race_state_id",
                                 "Age", "Probability"])


for i, row in df.iterrows():
    for p in range(int(row['Age_low']), int(row['Age_high'])):
        new_row = {
        "Model_id": row['Model_id'],
        "Model_name": row['Model_name'],
        "To_state_id": row['To_state_id'],
        "To_state_name": row['To_state_name'],
        "Sex_state_id": row['Sex_state_id'],
		"Race_state_id": row['Race_state_id'],
		"Age": p,
		"Probability": row['Probability']
		}
        outputdf = outputdf.append(new_row, ignore_index=True)

outputdf.to_csv("ras_output.csv")
