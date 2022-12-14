#!/usr/bin/env python
'''
@FileName    : example1.py
@Description : A example of DAG graph visualization.
@Date        : 2022/12/14 17:30:37
@Author      : Wiger
@version     : 1.0
'''
# Import the required libraries
import networkx as nx
import matplotlib.pyplot as plt

# Create an empty graph with no nodes and no edges
g = nx.DiGraph()

# Add nodes to the graph
g.add_node(1)
g.add_node(2)
g.add_node(3)
g.add_node(4)
g.add_node(5)
g.add_node(6)

# Add edges to the graph
g.add_edge(1, 2)
g.add_edge(1, 3)
g.add_edge(3, 2)
g.add_edge(4, 2)
g.add_edge(5, 5)
g.add_edge(6, 3)
g.add_edge(5, 6)

# Draw the graph
nx.draw(g, with_labels=True)
plt.show()
