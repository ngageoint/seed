import json
import os
import shutil
import sys
from PIL import Image

def rotate_image(source_image_path, rotate_degrees, output_dir):
    """Rotates an image a specified number of degrees"""
    product_image_name =  f'ROTATED_{rotate_degrees}_{os.path.basename(source_image_path)}'

    # Open source image
    original_image = Image.open(source_image_path)

    # Rotate it specified number of degrees
    rotated_image = original_image.rotate(int(rotate_degrees), 0, 1)
    rotated_image.save(product_image_name)

    # Move saved image to output directory
    product_image_path = os.path.join(output_dir, product_image_name)
    if not os.path.exists(output_dir):
        os.makedirs(output_dir)
    shutil.move(product_image_name, product_image_path)

    return product_image_name

def write_results_manifest(output_dir, product_image):

    with open(os.path.join(output_dir, 'results_manifest.json'), 'w+') as fout:
        json_string = json.dumps({'output_data': [{'name': 'ROTATED_IMAGE','path': product_image}]})
        fout.write(json_string)
        print(json_string)

if __name__ == "__main__":

    source_image_path = sys.argv[1]
    rotate_degrees = sys.argv[2]
    output_dir = sys.argv[3]

    product_image_path = rotate_image(source_image_path, rotate_degrees, output_dir)
    write_results_manifest(output_dir, product_image_path)