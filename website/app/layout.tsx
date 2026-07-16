import type { Metadata } from 'next'
import { GeistSans } from 'geist/font/sans'
import { GeistMono } from 'geist/font/mono'
import './globals.css'
import { Navbar } from '@/components/Navbar'
import { Footer } from '@/components/Footer'

export const metadata: Metadata = {
  title: {
    default: 'Orbita — self-hosted PaaS + Grit Cloud',
    template: '%s · Orbita',
  },
  description:
    'A self-hosted, multi-tenant PaaS in a single Go binary — and the control plane for Grit Cloud. Secure a server and deploy a full Grit app in two commands.',
  metadataBase: new URL('https://orbita.example.com'),
  openGraph: {
    title: 'Orbita — self-hosted PaaS + Grit Cloud',
    description: 'One VPS. Many clients. Full isolation. Deploy a full Grit app in two commands.',
    type: 'website',
  },
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className={`dark ${GeistSans.variable} ${GeistMono.variable}`}>
      <body className="bg-bgDark antialiased">
        <Navbar />
        {children}
        <Footer />
      </body>
    </html>
  )
}
