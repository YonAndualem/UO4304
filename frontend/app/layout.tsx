import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import { IdentityProvider } from "@/contexts/IdentityContext";
import { DemoModeProvider } from "@/contexts/DemoModeContext";
import { NavBar } from "@/components/NavBar";
import { DemoBanner } from "@/components/DemoBanner";
import { RoleBar } from "@/components/RoleBar";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "Trade License Portal",
  description: "Enterprise trade license workflow management",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={`${geistSans.variable} ${geistMono.variable} h-full antialiased`}
    >
      <body className="min-h-full flex flex-col bg-gray-50">
        <IdentityProvider>
          <DemoModeProvider>
            <NavBar />
            <DemoBanner />
            <RoleBar />
            <main className="flex-1">{children}</main>
          </DemoModeProvider>
        </IdentityProvider>
      </body>
    </html>
  );
}
