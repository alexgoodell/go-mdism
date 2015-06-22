


import os
import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
from bokeh.plotting import figure, output_file, show
from bokeh.palettes import brewer
palette = brewer["Spectral"]

os.chdir('/Applications/XAMPP/xamppfiles/htdocs/ghd_model_go/gocode/src/github.com/alexgoodell/go-mdism/tmp/')

df = pd.read_csv('state_populations.csv', header=0, skip_blank_lines=1)

df = df[(df.Model_id == 0)]
df = df.drop('Model_id', axis=1, level=None, inplace=False)
df = df.drop('Id', axis=1, level=None, inplace=False)


# create a new plot
p = figure(
   tools="pan,box_zoom,reset,save",
   title="state populations",
   x_axis_label='cycles', y_axis_label='number of people'
)

state_ids = [ 0 , 1, 2, 3, 4, 5, 6, 7, 8, 9]
x = []
y = []
state_names =[ "Unitialized1", "Unitialized2", "No NAFLD", "Steatosis", "NASH", "Cirrhosis", "HCC", "Liver death", "Natural death", "Other death" ]


for i in state_ids:
    x.append(df[df.State_id == i].Cycle_id)
    y.append(df[df.State_id == i].Population)
    p.line(x[i], y[i], legend=state_names[i], color=palette[10][i], line_width=3)    

output_file("log_lines.html")
show(p)


df.State_id = df.State_id.replace([ 0 , 1, 2, 3, 4, 5, 6, 7, 8, 9], )

final = df.pivot_table(index='State_id',columns='Cycle_id')



os.chdir('/Applications/XAMPP/xamppfiles/htdocs/ghd_model_go/gocode/src/github.com/alexgoodell/go-mdism/analysis_output')


for i, row in  final.iterrows():
    print row['Cycle_id']
   print row['Cycle_id'], row['Population']
    

# show the res





