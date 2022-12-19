#!/usr/bin/env python
'''
@FileName    : example3.py
@Description : A example of DAG graph visualization.
@Date        : 2022/12/18 23:18:20
@Author      : Wiger
@version     : 1.0
'''
# Import the required libraries
import networkx as nx
import matplotlib.pyplot as plt
import random

# Create an empty graph
G = nx.DiGraph()

# Add 10 nodes
for i in range(70):
    G.add_node(i)

# Add directed edges
for i in range(50):
    G.add_edge(random.randint(0,70), random.randint(0,70))

# Draw the graph
nx.draw(G, with_labels=True)
plt.show()
