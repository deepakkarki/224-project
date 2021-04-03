import os

base = os.path.join(os.path.dirname(__file__), '../sample_htdocs')


with open(base + "/index.html", "rb") as f:
  root_index = f.read()

with open(base + "/subdir1/index.html", "rb") as f:
  subdir1_index = f.read()

with open(base + "/subdir1/subdir11/maoyo.giaogiao", "rb") as f:
  maoyo_giaogiao = f.read()

with open(base + "/kitten.jpg", "rb") as f:
  kitten = f.read()
