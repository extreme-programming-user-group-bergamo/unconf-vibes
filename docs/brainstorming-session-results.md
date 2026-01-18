# Brainstorming Session: UNCONF CLI User Experience

**Date**: January 18, 2026  
**Facilitator**: Mary (Business Analyst)  
**Topic**: User Experience / TUI Design for UNCONF CLI  
**Goal**: Broad exploration of UX directions  

## Session Parameters

| Parameter | Value |
|-----------|-------|
| **Focus Area** | TUI design, interactions, user flows |
| **Constraints** | Broad terminal compatibility, all dev levels, Open Space/Unconference format |
| **Approach** | Analyst-recommended techniques |
| **Duration** | ~60 minutes target |

---

## Executive Summary

**Session Topic**: User Experience / TUI Design for UNCONF CLI  
**Duration**: ~45 minutes  
**Techniques Used**: Role Playing, What If Scenarios, Analogical Thinking, SCAMPER  

### Key Outcomes

- **34 ideas generated** across 4 techniques
- **13 MVP features** identified and prioritized
- **5 design principles** established to guide development
- **4 personas** explored: First-Timer, Organizer, Returning Attendee, Privacy-Conscious

### Top 3 Insights

1. **Dual-mode architecture**: TUI for discovery and exploration, CLI for fast execution
2. **Privacy as a feature**: Visibility tiers (Private/Protected/Public) are core, not optional
3. **Progressive disclosure**: Essential for serving both first-timers and power users

### Recommended Next Steps

1. Create detailed user flows for the MVP features
2. Design wireframes/mockups for TUI screens (conference list, room selection, attendee view)
3. Define the CLI command structure and help text tone
4. Establish the privacy model and data visibility rules
### Reconciliation with MVP.md

The brainstorming session has been reconciled with the existing [MVP.md](../MVP.md) document:

**Added to MVP scope from brainstorming:**
- Fuzzy search for conferences
- First-Timer Guide command
- Calendar .ics file generation (in addition to email)

**Confirmed alignment:**
- Privacy tiers (public/private)
- Context system (checkout)
- TUI for discovery, CLI for action
- Social discovery (attendee visibility)
- Organizer dashboard features
- Power user shortcuts (--force, direct room booking)

**Already in MVP.md, now categorized:**
- Hotel email automation → MVP essential
- Dietary/accessibility preferences → MVP essential
- Cancellation flow → MVP P2
- CSV export → MVP P2
---

## Technique 1: Role Playing (Stakeholder Perspectives)

**Rationale**: Understanding different users ensures the UX works for everyone attending an unconference.

### Ideas Generated

#### Persona 1: The First-Timer
*Someone attending their first Open Space conference, decent developer but not a terminal power-user*

**Pain Points Identified:**
- Too much information at once would overwhelm/scare them
- Don't know what Open Space format means or conference norms
- Uncertainty about the booking process

**UX Ideas:**
- 🔍 **Fuzzy search for conferences** — Simple "search by name" with fuzzy-matching to discover events
- 📖 **"First-Timer Guide" command** — Dedicated help explaining the conference style and format
- 💬 **Friendly, direct language** — Don't assume prior knowledge; warm and welcoming tone
- 📊 **Progressive disclosure** — Don't dump all info at once; reveal as needed

#### Persona 2: The Organizer
*Conference organizer managing the event setup, attendees, and logistics*

**Pain Points Identified:**
- Need quick access to information or they'll switch to other tools
- Managing capacity limits and special requests
- Keeping track of who's registered and their preferences

**UX Ideas:**
- 📋 **Conference listing dashboard** — Easy view of all their unconferences
- 🔗 **Shareable detail links** — Quick way to share conference info with attendees
- 👥 **Capacity management** — Set and track max attendees
- 📊 **Attendee overview** — See registered users with preferences (dietary, special requests)
- ⚡ **Quick-access commands** — No friction to get critical info

#### Persona 3: The Returning Attendee / Community Regular
*Experienced attendee who's been to multiple unconferences and knows the community*

**Pain Points Identified:**
- Don't want to wade through explanations they already know
- Want to see who else is going (social discovery)

**UX Ideas:**
- 👀 **Attendee list (opt-in visibility)** — See who's going if they marked themselves public
- ⏩ **Fast-track booking** — Skip explanations, go straight to room selection
- 🔄 **"I know this" shortcuts** — Detect returning users or offer express mode

#### Persona 4: The Privacy-Conscious Attendee
*Someone who wants to attend but is cautious about visibility and information sharing*

**Pain Points Identified:**
- Don't want their attendance publicly visible
- Need to control what information is shared
- Still need organizers to know they're coming

**UX Ideas:**
- 🔒 **Visibility tiers** — Private / Protected / Public attendance modes
- 📉 **Bare-minimum sharing** — Only essential info visible to other attendees
- 👁️ **Organizer-only view** — Organizers can always see attendance for logistics
- ⚙️ **Granular privacy controls** — Clear UI to understand what's shared with whom

### Technique Summary
**Ideas Generated**: 16  
**Key Themes**: Progressive disclosure, Express vs Guided modes, Privacy tiers, Social discovery, Quick access

---

## Technique 2: What If Scenarios

**Rationale**: Challenge assumptions about terminal UIs to discover innovative approaches.

### Ideas Generated

#### Scenario 1: What if NO visual TUI? (Pure CLI)
**Constraint**: Remove all Bubble Tea TUI — just Unix commands and text output

**What We'd Lose:**
- Clarity when presenting lots of information
- Guided experience for first-timers — info gets lost in text walls

**What We'd Gain:**
- ⚡ Speed of operations
- 📋 Scriptability / automation
- 🔧 Works everywhere, even minimal terminals

**Insight → Hybrid Approach:**
- 💡 **TUI for discovery, CLI for action** — Use TUI when exploring (rooms, attendees) but allow CLI shortcuts for power users who know what they want

#### Scenario 2: What if ONE command could book everything?
**Concept**: `unconf book socrates-italia --room 204 --roommate @marco`

**Who Benefits:**
- ⚡ Power users who know exactly what they want
- 🔄 Recurring attendees rebooking familiar setups

**Challenges:**
- 🏨 External hotel booking interactions might still be needed
- ❓ Requires knowing room numbers, roommate handles in advance

**Insight:**
- 💡 **Fast-path for the certain** — Offer one-liner booking as an "expert mode" but default to guided flow

#### Scenario 3: What if room selection was like a VIDEO GAME?
**Concept**: Avatar-based navigation, retro game feel, chat bubbles with other attendees

**Verdict: ❌ Too much**
- Distracts from core goal
- UNCONF's purpose is to *enrich in-person interactions*, not replace them with virtual ones
- Game-like elements could feel gimmicky for professional context

**Insight:**
- 💡 **Functional delight over entertainment** — TUI should be satisfying to use, not entertaining. The "fun" happens at the conference, not in the CLI

#### Scenario 4: What if OFFLINE-FIRST?
**Concept**: Cache data locally, browse offline, sync when connected

**Verdict: 🟡 Nice-to-have, not required**
- Not critical for organizers (they need real-time data)
- Could be useful for attendees traveling to venue
- Low priority for MVP

**Insight:**
- 💡 **Graceful degradation** — Show cached data when offline but clearly indicate staleness

#### Scenario 5: What if TERMINAL NOTIFICATIONS?
**Concept**: Push notifications for room availability, roommate requests, deadlines

**Verdict: 🟡 Useful but complex**
- Push notifications to terminal are technically complex
- Email notifications for MVP — simpler, broader reach
- Terminal notifications could be a v2 feature

**Insight:**
- 💡 **Email as MVP notification channel** — Then expand to terminal/webhook notifications later

### Technique Summary
**Scenarios Explored**: 5  
**Key Principles Discovered**: TUI/CLI hybrid, Expert mode, Functional delight, Graceful degradation, Email-first notifications

---

## Technique 3: Analogical Thinking

**Rationale**: Draw inspiration from successful UX patterns in non-CLI contexts.

### Ideas Generated

#### Analogy 1: Airline Seat Selection
**Verdict: ❌ Out of scope**
- Physical representation of room layout not required
- UNCONF is simpler: just know room size and who's sharing
- Hotel rooms aren't positioned like airplane seats

**Insight:**
- 💡 **Simplicity over visualization** — Room attributes (size, occupancy, roommate) matter more than physical layout

#### Analogy 2: Discord/Slack Community Joining
**Source inspiration**: Joining a Discord server or Slack workspace

**What translates well:**
- 🔍 **Discovery** — Search/browse communities, invite links
- 📋 **Quick orientation** — See name, description, member count before joining
- 🎭 **Identity setup** — Choose display name on join
- 🔒 **Privacy controls** — Granular visibility settings

**UX Ideas for UNCONF:**
- 📨 **Invite link sharing** — `unconf join <invite-code>` for quick onboarding
- 👀 **Conference preview** — See details and attendee count before committing
- 🏷️ **Display name per conference** — Use different names at different events
- ⚙️ **Privacy on join** — Set visibility as part of the join flow

#### Analogy 3: Git/kubectl Context Switching
**Source inspiration**: Context-based workflows in developer tools

**Keep it simple for MVP:**
- Current context always visible in TUI
- Clear indication of which conference you're operating on

**UX Ideas for UNCONF:**
- 🏷️ **Persistent context indicator** — Always show current conference in TUI header/footer
- ⚠️ **Context confirmation for destructive actions** — "You're about to cancel booking for SoCraTes Italia"

#### Analogy 4: Airbnb/Booking.com Social Proof
**Source inspiration**: Social proof and urgency indicators

**UX Ideas for UNCONF:**
- 📊 **Attendance momentum** — "5 people booked today", "Last booking: 2 hours ago"
- 👥 **Occupancy indicators** — "3 of 4 spots filled in Room 204"
- 🏆 **Past event highlights** — Returning attendees or "This is Marco's 3rd year"
- ⏰ **Subtle urgency** — "Booking closes in 5 days" (not aggressive)

#### Analogy 5: Eventbrite/Meetup Event Discovery
**Source inspiration**: Event platform discovery and engagement patterns

**UX Ideas for UNCONF:**
- 🔍 **Event discovery** — Browse/search unconferences with filters
- 💚 **Interest states** — "Interested" vs "Going" to gauge demand before committing
- 📈 **Historical popularity** — "Last year: 45 attendees" to show track record
- 🔥 **Edition appeal** — See current interest level for upcoming edition

### Technique Summary
**Analogies Explored**: 5 (1 rejected as out-of-scope)  
**Key Patterns Borrowed**: Community joining flow, Context indicators, Social proof, Interest/Going states

---

## Technique 4: SCAMPER Method

**Rationale**: Systematically improve and refine the concepts we've generated.

### Ideas Generated

#### S = Substitute
**Assessment:** Limited room for substitution — simplicity is the goal
- Core flows should remain straightforward
- Don't over-engineer what works

#### C = Combine
**Promising combinations identified:**

- 📋 **Conference Info + Attendee List** — Single view to gauge appeal
  - See event details AND who's going in one place
  - Helps decision-making for first-timers
  
- ✅ **Booking Confirmation + Calendar Invite** — Seamless follow-through
  - Generate .ics file or calendar link on confirmation
  - Keep attendees informed without extra steps

#### A = Adapt
*Covered in Analogical Thinking technique — no additional adaptations identified*

#### M = Modify / Magnify / Minimize

**MAGNIFY** 🔍
- Social discovery — Make finding events and seeing who's going prominent and engaging
- Community aspect — Highlight the "who" not just the "what"

**MINIMIZE** 📉
- Steps to book — Reduce friction, fewer clicks/commands to complete a booking
- Information overload — Progressive disclosure, show only what's needed when needed

#### P = Put to Other Uses

**Future expansion idea:**
- 📊 **Historical data for event planning** — Past attendance, room preferences, dietary needs help organizers set up new events faster
- Template from previous editions
- Predictive capacity planning based on interest signals

#### E = Eliminate
**Assessment:** Current feature set is appropriately scoped
- No unnecessary features identified for removal
- Focus on doing fewer things well

#### R = Rearrange / Reverse
**Assessment:** Current flow order is logical
- No rearrangements needed

### Technique Summary
**SCAMPER Results:**
- S: Keep simple ✓
- C: Info+Attendees, Confirmation+Calendar ★
- A: Covered in analogies
- M: Magnify discovery, minimize friction ★
- P: Historical data for organizers ★
- E: Nothing to cut
- R: Current order works

---

## Synthesis & Prioritization

### All Ideas Generated (Consolidated)

| # | Idea | Source | Category |
|---|------|--------|----------|
| 1 | Fuzzy search for conferences | Role Playing | Discovery |
| 2 | First-Timer Guide command | Role Playing | Onboarding |
| 3 | Friendly, direct language | Role Playing | UX Tone |
| 4 | Progressive disclosure | Role Playing | Information Architecture |
| 5 | Conference listing dashboard (organizers) | Role Playing | Organizer Tools |
| 6 | Shareable detail links | Role Playing | Sharing |
| 7 | Capacity management | Role Playing | Organizer Tools |
| 8 | Attendee overview with preferences | Role Playing | Organizer Tools |
| 9 | Quick-access commands | Role Playing | Efficiency |
| 10 | Attendee list (opt-in visibility) | Role Playing | Social Discovery |
| 11 | Fast-track booking / Express mode | Role Playing | Efficiency |
| 12 | Visibility tiers (Private/Protected/Public) | Role Playing | Privacy |
| 13 | Bare-minimum sharing | Role Playing | Privacy |
| 14 | Granular privacy controls | Role Playing | Privacy |
| 15 | TUI for discovery, CLI for action | What If | Architecture |
| 16 | One-command expert booking | What If | Efficiency |
| 17 | Functional delight over entertainment | What If | Design Philosophy |
| 18 | Graceful offline degradation | What If | Resilience |
| 19 | Email as MVP notification channel | What If | Notifications |
| 20 | Invite link sharing | Analogies | Onboarding |
| 21 | Conference preview before committing | Analogies | Discovery |
| 22 | Display name per conference | Analogies | Identity |
| 23 | Privacy settings on join flow | Analogies | Privacy |
| 24 | Persistent context indicator | Analogies | Navigation |
| 25 | Context confirmation for destructive actions | Analogies | Safety |
| 26 | Attendance momentum indicators | Analogies | Social Proof |
| 27 | Occupancy indicators | Analogies | Social Proof |
| 28 | Past event highlights | Analogies | Social Proof |
| 29 | Subtle urgency (deadlines) | Analogies | Engagement |
| 30 | Interest/Going states | Analogies | Engagement |
| 31 | Historical popularity display | Analogies | Discovery |
| 32 | Conference Info + Attendee List combo view | SCAMPER | UX Combination |
| 33 | Booking Confirmation + Calendar Invite | SCAMPER | UX Combination |
| 34 | Historical data for event planning | SCAMPER | Organizer Tools |

**Total Ideas: 34**

---

### Idea Categorization

#### 🟢 Immediate Opportunities (MVP)
*Ready to implement now — core to the product*

| Priority | Idea | Rationale |
|----------|------|-----------|
| P1 | Fuzzy search for conferences | Core discovery feature |
| P1 | First-Timer Guide command | Essential for onboarding |
| P1 | Friendly, direct language | Sets the tone for all UX |
| P1 | Progressive disclosure | Prevents overwhelming users |
| P1 | Visibility tiers (Private/Protected/Public) | Privacy is a core promise |
| P1 | Persistent context indicator | Navigation clarity |
| P2 | Fast-track booking / Express mode | Power user efficiency |
| P2 | Booking Confirmation + Calendar Invite | Polish that delights |
| P2 | Conference preview before committing | Builds confidence |
| P2 | Attendee list (opt-in visibility) | Social discovery core feature |
| P2 | Organizer: Conference listing dashboard | Essential for conference setup |
| P2 | Organizer: Attendee overview with preferences | Core organizer functionality |
| P2 | Organizer: Capacity management | Required for practical use |

#### 🟢 Additional MVP Features (from MVP.md reconciliation)
*Already specified in MVP.md — incorporated into scope*

| Priority | Idea | Source |
|----------|------|--------|
| P1 | Hotel email automation | MVP.md — core automation feature |
| P1 | Dietary/accessibility preferences (account setup) | MVP.md — account flow |
| P1 | Room type filtering (`--type single|double|triple`) | MVP.md — room discovery |
| P2 | Cancellation flow (`cancel` command) | MVP.md — booking lifecycle |
| P2 | Export to CSV (organizers) | MVP.md — data extraction |
| P2 | Scripting mode (`--json` output) | MVP.md — power user/automation |

#### 🟡 Future Innovations (V2)
*Valuable enhancements for post-MVP releases*

| Idea | Notes |
|------|-------|
| Email notifications | Complement calendar with email updates |
| Interest/Going states | Gauge demand before committing |
| Social proof indicators | Momentum, occupancy signals |
| Historical popularity display | Past attendance data |
| One-command expert booking | `unconf book <conf> --room X --roommate Y` |
| Shareable invite links | Quick onboarding codes |
| Display name per conference | Different identity per event |
| TUI for discovery, CLI for action | Hybrid architecture refinement |
| Context confirmation for destructive actions | Safety guardrails |
| Subtle urgency indicators | Booking deadlines |
| Display name per conference | Different identity per event |
| 3-tier privacy (Protected) | Extend beyond public/private |
| .ics calendar file generation | Enhance email-only confirmation |
| Invite link sharing | Quick onboarding codes |

#### 🔵 Moonshots (Long-term Vision)
*Ambitious ideas for significant future investment*

| Idea | Notes |
|------|-------|
| Historical data for event planning | Predictive organizer tools |
| Graceful offline mode | Cache and sync |
| Past event highlights | "Marco's 3rd year" badges |
| Template conferences from past editions | One-click setup for recurring events |

---

## Key Design Principles

Based on our brainstorming session, these principles should guide all UX decisions:

### 1. Progressive Disclosure
> Reveal information as needed, don't overwhelm users upfront.

- First-timers see the essentials; power users can dig deeper
- Context-sensitive help over comprehensive documentation
- Default to simple, offer complexity

### 2. TUI for Discovery, CLI for Action
> Visual interfaces for exploration, command-line for execution.

- Interactive TUI when browsing rooms, attendees, conferences
- One-liner commands for users who know what they want
- Both paths reach the same destination

### 3. Privacy by Design
> Visibility controls are upfront, not buried in settings.

- Ask about privacy during onboarding flow
- Clear mental model: Private / Protected / Public
- Organizers always see what they need; attendees control peer visibility

### 4. Functional Delight
> The UX should be satisfying to use, not entertaining.

- Smooth, responsive interactions
- Clear feedback for every action
- The "fun" happens at the conference, not in the CLI

### 5. Magnify Community, Minimize Friction
> Highlight the social, reduce the bureaucratic.

- Make "who's going?" easy to discover
- Minimize steps between "I want to go" and "I'm booked"
- Social proof builds excitement; friction kills it

---

## Action Planning

### Recommended Next Steps

| # | Action | Owner | Notes |
|---|--------|-------|-------|
| 1 | Create detailed user flows for MVP features | UX/Dev | Start with booking flow and organizer setup |
| 2 | Design TUI wireframes | UX | Conference list, room selection, attendee view |
| 3 | Define CLI command structure | Dev | Include help text tone guide |
| 4 | Establish privacy model | Architect | Private/Protected/Public rules |
| 5 | Write First-Timer Guide content | Content | Friendly, concise, welcoming |

### Questions for Further Exploration

- How will the actual hotel booking integration work? (External system? API?)
- What's the relationship between UNCONF and conference registration?
- How do organizers initially set up room inventory?
- What happens when roommate preferences conflict?

---

## Session Reflection

### What Worked Well
- Role Playing revealed key persona differences (first-timer vs power user)
- What If Scenarios helped establish boundaries (no game-like UI)
- Analogical Thinking brought proven patterns from Discord/Airbnb/Eventbrite
- SCAMPER confirmed simplicity focus while surfacing combination opportunities

### Areas for Future Sessions
- Deep dive on organizer workflows
- Roommate matching/discovery mechanics
- Error handling and edge cases in booking flow
- Accessibility considerations for TUI

---

*Session facilitated by Mary (Business Analyst) using BMAD Method*  
*Document generated: January 18, 2026*

