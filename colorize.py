import sys

import tensorflow as tf
import skimage.transform
from skimage.io import imsave, imread

def load_image(path):
	img = imread(path)
	# crop image from center
	short_edge = min(img.shape[:2])
	yy = int((img.shape[0] - short_edge) / 2)
	xx = int((img.shape[1] - short_edge) / 2)
	crop_img = img[yy : yy + short_edge, xx : xx + short_edge]
	# resize to 224, 224
	img = skimage.transform.resize(crop_img, (224, 224))
	# desaturate image
	return (img[:,:,0] + img[:,:,1] + img[:,:,2]) / 3.0

img_gray = load_image(sys.argv[1]).reshape(1, 224, 224, 1)

with open("/colorize/colorize.tfmodel", mode='rb') as f:
	fileContent = f.read()

graph_def = tf.GraphDef()
graph_def.ParseFromString(fileContent)
grayscale = tf.placeholder("float", [1, 224, 224, 1])
tf.import_graph_def(graph_def, input_map={ "grayscale": grayscale }, name='')

oname = sys.argv[2]

with tf.Session() as sess:
	inferred_rgb = sess.graph.get_tensor_by_name("inferred_rgb:0")
	inferred_batch = sess.run(inferred_rgb, feed_dict={ grayscale: img_gray })
	imsave(oname, inferred_batch[0])
	print "saved {}".format(oname)

if len(sys.argv) >= 3:
	imsave(sys.argv[3], img_gray)
