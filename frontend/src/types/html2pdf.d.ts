import "html2pdf.js";

declare module "html2pdf.js" {
  interface Html2PdfOptions {
    pagebreak?: {
      mode?: "avoid-all" | "css" | "legacy" | Array<"avoid-all" | "css" | "legacy">;
      before?: string | string[];
      after?: string | string[];
      avoid?: string | string[];
    };
  }
}