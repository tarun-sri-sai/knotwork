import type { Metadata } from "next";
import { JetBrains_Mono } from "next/font/google";
import "./globals.css";
import Header from "@/app/Header";

const monoFont = JetBrains_Mono({
  variable: "--font-mono",
  subsets: ["latin"],
  weight: ["300", "400"],
});

export const metadata: Metadata = {
  title: "Knotwork",
  description: "Progress tracking made faster, simpler, and better.",
};

const RootLayout = ({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) => {
  return (
    <html lang="en">
      <body className={`${monoFont.variable} antialiased`}>
        <Header />
        <main className="text-sm">{children}</main>
      </body>
    </html>
  );
};

export default RootLayout;
