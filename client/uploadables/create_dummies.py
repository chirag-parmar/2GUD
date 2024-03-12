for i in range(0, 1000):
    with open(str(i) + ".txt", "w") as f:
        f.write(str(i)*i)
    f.close()