import InfoPageView, { infoMetadata } from "../info-page";

export function generateMetadata() {
  return infoMetadata("contact");
}

export default function Page() {
  return <InfoPageView slug="contact" />;
}
