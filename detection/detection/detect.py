from PIL import ImageFile
import time
from ultralytics import YOLO


model = YOLO('yolov8n.pt')


def predict(image: ImageFile.ImageFile) -> list[dict]:
    results = model.predict(source=image, imgsz=320, conf=0.1)
    detections = []
    for result in results:
        for box in result.boxes.data.tolist():
            x1, y1, x2, y2, score, class_id = box
            cls = model.names[int(class_id)]
            detections.append({
                'class': cls,
                'score': float(score),
                'box': [x1, y1, x2, y2]
            })
    time.sleep(1)
    return detections
