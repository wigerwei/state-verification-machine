#!/usr/bin/env python
'''
@FileName    : example2.py
@Description : A example of DAG graph visualization.
@Date        : 2022/12/14 17:34:27
@Author      : Wiger
@version     : 1.0
'''
# Import the required libraries
import networkx as nx
import matplotlib.pyplot as plt

# Create an empty graph
G = nx.DiGraph()

# Add 10 nodes
for i in range(10):
    G.add_node(i)

# Add directed edges
G.add_edge(0, 1)
G.add_edge(1, 2)
G.add_edge(2, 1)
G.add_edge(3, 4)
G.add_edge(3, 6)
G.add_edge(3, 9)
G.add_edge(4, 5)
G.add_edge(5, 6)
G.add_edge(7, 8)
G.add_edge(8, 9)

# Draw the graph
nx.draw(G, with_labels=True)
plt.show()
