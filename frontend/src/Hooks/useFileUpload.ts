import { useState } from "react";
import type { FileItem } from "../types";

export function useFileUpload() {
  const [fileList, setFileList] = useState<FileItem[]>([]);
  const [previewImage, setPreviewImage] = useState<string | null>(null);

  const openPreview = (url: string) => setPreviewImage(url);
  const closePreview = () => setPreviewImage(null);

  return { fileList, setFileList, previewImage, openPreview, closePreview };
}