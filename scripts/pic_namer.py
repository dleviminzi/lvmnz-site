import os
import uuid
from datetime import datetime

directory_path = "./static/photos/new"

for filename in os.listdir(directory_path):
    if filename.lower().endswith('.jpeg'):
        full_path = os.path.join(directory_path, filename)
        stat_info = os.stat(full_path)
        birth_time = stat_info.st_birthtime
        formatted_time = datetime.fromtimestamp(birth_time).strftime('%Y-%m-%d')
        
        # avoid overwriting files created on the same day
        unique_id = uuid.uuid4()
        
        new_filename = f"{formatted_time}_{unique_id}.jpeg"
        
        new_full_path = os.path.join(directory_path, new_filename)
        
        os.rename(full_path, new_full_path)
        print(f"Renamed {filename} to {new_filename}")

print("All JPEG files have been renamed.")
