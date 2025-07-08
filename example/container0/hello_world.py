import numpy as np
import torch

if __name__ == "__main__":
    # Create a random tensor
    tensor = torch.rand(3, 3)
    
    # Convert to numpy array
    np_array = tensor.numpy()
    
    print("Using PyTorch and NumPy")
    print("Random Tensor:")
    print(tensor)
    print("\nConverted Numpy Array:")
    print(np_array)