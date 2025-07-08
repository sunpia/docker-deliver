import numpy as np
import torch
import pandas as pd

if __name__ == "__main__":
    # Create a random tensor
    tensor = torch.rand(3, 3)
    
    # Convert to numpy array
    np_array = tensor.numpy()
    
    print("Using PyTorch, Pandas, and NumPy")
    print("Random Tensor:")
    print(tensor)
    print("\nConverted Numpy Array:")
    print(np_array)