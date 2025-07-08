import numpy as np
import torch
import matplotlib.pyplot as plt
import pandas as pd
from sklearn import datasets


if __name__ == "__main__":
    # Create a random tensor
    tensor = torch.rand(3, 3)
    
    # Convert to numpy array
    np_array = tensor.numpy()
    
    print("Using PyTorch, NumPy, Matplotlib, and Scikit-learn")
    print("Random Tensor:")
    print(tensor)
    print("\nConverted Numpy Array:")
    print(np_array)