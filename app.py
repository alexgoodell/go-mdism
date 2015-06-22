'''
This script is loosely based on the bokeh spectogram example,
but is much simpler:

    https://github.com/bokeh/bokeh/tree/master/examples/embed/spectrogram

This creates a simple form for generating polynomials of the form y = x^2.

This is done using a form that has a method of GET, allowing you to share the
graphs you create with your friends though the link!

You should know at least the basics of Flask to understand this example
'''


import flask
import os
import pandas as pd
import numpy as np
from bokeh.plotting import figure, output_file, show
from bokeh.palettes import brewer
from bokeh.embed import components
from bokeh.resources import INLINE
from bokeh.templates import RESOURCES
from bokeh.util.string import encode_utf8


app = flask.Flask(__name__)


selector = "Please select chain: "

models_df = pd.read_csv('inputs/example/models.csv', header=0, skip_blank_lines=1)
for index, row in models_df.iterrows():
    selector += "<a href=\"/model/" + str(row["Id"]) + "\">" + row["Name"] +"</a> "



@app.route("/")
def base():
    reutn "hello"


@app.route("/model/<int:model_id>")
def select_model(model_id):

    palette = brewer["Spectral"]
    state_pop_df = pd.read_csv('tmp/state_populations.csv', header=0, skip_blank_lines=1)
    state_pop_df = state_pop_df[(state_pop_df.Model_id == model_id)]
    state_pop_df = state_pop_df.drop('Model_id', axis=1, level=None, inplace=False)
    state_pop_df = state_pop_df.drop('Id', axis=1, level=None, inplace=False)

    # create a new plot
    fig = figure(
        tools="pan,box_zoom,reset,save",
        title="state populations",
        x_axis_label='cycles', y_axis_label='number of people'
    )

    models_df = pd.read_csv('inputs/example/models.csv', header=0, skip_blank_lines=1)
    states_df = pd.read_csv('inputs/example/states.csv', header=0, skip_blank_lines=1)

    state_ids = list(states_df[states_df.Model_id == model_id].Id)
    x = []
    y = []
    state_names = list(states_df[states_df.Model_id == model_id].Name)

    for i, s in enumerate(state_ids):
        x.append(state_pop_df[state_pop_df.State_id == s].Cycle_id)
        y.append(state_pop_df[state_pop_df.State_id == s].Population)
        fig.line(x[i], y[i], legend=state_names[i], color=palette[10][i], line_width=3)



    # Get all the form arguments in the url with defaults
    color = "Black"

    # Configure resources to include BokehJS inline in the document.
    # For more details see:
    # http://bokeh.pydata.org/en/latest/docs/reference/resources_embedding.html#module-bokeh.resources
    plot_resources = RESOURCES.render(
        js_raw=INLINE.js_raw,
        css_raw=INLINE.css_raw,
        js_files=INLINE.js_files,
        css_files=INLINE.css_files,
    )

    # For more details see:
    # http://bokeh.pydata.org/en/latest/docs/user_guide/embedding.html#components
    script, div = components(fig, INLINE)
    html = flask.render_template(
        'embed.html',
        plot_script=script, plot_div=div, plot_resources=plot_resources,
        color=color, selector=selector
    )
    return encode_utf8(html)


def main():
    app.debug = True
    app.run()

if __name__ == "__main__":
    main()
