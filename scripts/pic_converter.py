from PIL import Image
import os
import shutil

def convert_jpeg_to_webp(jpeg_path, webp_path):
    # Open the JPEG image
    img = Image.open(jpeg_path)
    
    img.save(webp_path, 'WEBP', quality=70)

input_directory = "./raw"
output_directory = os.path.join(input_directory, "original")

if not os.path.exists(output_directory):
    os.makedirs(output_directory)

files = os.listdir(input_directory)

for file_name in files:
    if file_name.lower().endswith('.jpg') or file_name.lower().endswith('.jpeg'):
        jpeg_path = os.path.join(input_directory, file_name)
        webp_name = os.path.splitext(file_name)[0] + '.webp'
        webp_path = os.path.join(input_directory, webp_name)

        convert_jpeg_to_webp(jpeg_path, webp_path)
        print(f"Converted {file_name} to {webp_name}")

        shutil.move(jpeg_path, os.path.join(output_directory, file_name))
        print(f"Moved {file_name} to the 'original' directory.")
