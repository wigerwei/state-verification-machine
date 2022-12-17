#!/usr/bin/env python
'''
@FileName    : dag-maker.py
@Description : Make a DAG of transactions according their statement.
@Date        : 2022/12/15 23:07:40
@Author      : Wiger
@version     : 1.0
'''
# Import dependencies
import networkx as nx
import matplotlib.pyplot as plt

G = nx.DiGraph()

G.add_node(1, amount=100, sender='Alice', receiver='Bob')
G.add_node(2, amount=50, sender='Charlie', receiver='Alice')
G.add_node(3, amount=75, sender='Bob', receiver='Charlie')

G.add_edge(1, 2)
G.add_edge(1, 3)
G.add_edge(2, 3)

# Draw the graph
nx.draw(G, with_labels=True)
plt.show()

# Function
def make_dag(txList, stateList):
    """
    @Description  : A function of building a DAG of transactions
    @Params       :
        txList: Transaction list.
        stateList: Statement of transactions list.
    @Returns      : A full DAG.
    """
